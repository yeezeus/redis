/*
Copyright The KubeDB Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controller

import (
	"context"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/apimachinery/pkg/eventer"
	validator "kubedb.dev/redis/pkg/admission"

	"github.com/appscode/go/log"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kutil "kmodules.xyz/client-go"
	dynamic_util "kmodules.xyz/client-go/dynamic"
)

func (c *Controller) create(redis *api.Redis) error {
	if err := validator.ValidateRedis(c.Client, c.ExtClient, redis, true); err != nil {
		c.recorder.Event(
			redis,
			core.EventTypeWarning,
			eventer.EventReasonInvalid,
			err.Error(),
		)
		log.Errorln(err)
		return nil // user error so just record error and don't retry.
	}

	if redis.Status.Phase == "" {
		rd, err := util.UpdateRedisStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), redis.ObjectMeta, func(in *api.RedisStatus) *api.RedisStatus {
			in.Phase = api.DatabasePhaseCreating
			return in
		}, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		redis.Status = rd.Status
	}

	// create Governing Service
	governingService := c.GoverningService
	if err := c.CreateGoverningService(governingService, redis.Namespace); err != nil {
		return err
	}

	// ensure ConfigMap for redis configuration file (i.e. redis.conf)
	if redis.Spec.Mode == api.RedisModeCluster {
		if err := c.ensureRedisConfig(redis); err != nil {
			return err
		}
	}

	// Ensure ClusterRoles for statefulsets
	if err := c.ensureRBACStuff(redis); err != nil {
		return err
	}

	// ensure database Service
	vt1, err := c.ensureService(redis)
	if err != nil {
		return err
	}

	// ensure database StatefulSet
	vt2, err := c.ensureRedisNodes(redis)
	if err != nil {
		return err
	}

	if vt1 == kutil.VerbCreated && vt2 == kutil.VerbCreated {
		c.recorder.Event(
			redis,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully created Redis",
		)
	} else if vt1 == kutil.VerbPatched || vt2 == kutil.VerbPatched {
		c.recorder.Event(
			redis,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully patched Redis",
		)
	}

	rd, err := util.UpdateRedisStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), redis.ObjectMeta, func(in *api.RedisStatus) *api.RedisStatus {
		in.Phase = api.DatabasePhaseRunning
		in.ObservedGeneration = redis.Generation
		return in
	}, metav1.UpdateOptions{})
	if err != nil {
		c.recorder.Eventf(
			redis,
			core.EventTypeWarning,
			eventer.EventReasonFailedToUpdate,
			err.Error(),
		)
		return err
	}
	redis.Status = rd.Status

	// ensure StatsService for desired monitoring
	if _, err := c.ensureStatsService(redis); err != nil {
		c.recorder.Eventf(
			redis,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorf("failed to manage monitoring system. Reason: %v", err)
		return nil
	}

	if err := c.manageMonitor(redis); err != nil {
		c.recorder.Eventf(
			redis,
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to manage monitoring system. Reason: %v",
			err,
		)
		log.Errorf("failed to manage monitoring system. Reason: %v", err)
		return nil
	}

	_, err = c.ensureAppBinding(redis)
	if err != nil {
		log.Errorln(err)
		return err
	}
	return nil
}

func (c *Controller) halt(db *api.Redis) error {
	if db.Spec.Halted && db.Spec.TerminationPolicy != api.TerminationPolicyHalt {
		return errors.New("can't halt db. 'spec.terminationPolicy' is not 'Halt'")
	}
	log.Infof("Halting Redis %v/%v", db.Namespace, db.Name)
	if err := c.haltDatabase(db); err != nil {
		return err
	}
	if err := c.waitUntilPaused(db); err != nil {
		return err
	}
	log.Infof("update status of Redis %v/%v to Halted.", db.Namespace, db.Name)
	if _, err := util.UpdateRedisStatus(context.TODO(), c.ExtClient.KubedbV1alpha1(), db.ObjectMeta, func(in *api.RedisStatus) *api.RedisStatus {
		in.Phase = api.DatabasePhaseHalted
		in.ObservedGeneration = db.Generation
		return in
	}, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func (c *Controller) terminate(redis *api.Redis) error {
	// If TerminationPolicy is "halt", keep PVCs,Secrets intact.
	if redis.Spec.TerminationPolicy == api.TerminationPolicyPause || redis.Spec.TerminationPolicy == api.TerminationPolicyHalt {
		if err := c.removeOwnerReferenceFromOffshoots(redis); err != nil {
			return err
		}
	} else {
		// If TerminationPolicy is "wipeOut", delete everything (ie, PVCs,Secrets,Snapshots).
		// If TerminationPolicy is "delete", delete PVCs and keep snapshots,secrets intact.
		// In both these cases, don't create dormantdatabase
		if err := c.setOwnerReferenceToOffshoots(redis); err != nil {
			return err
		}
	}

	if redis.Spec.Monitor != nil {
		if err := c.deleteMonitor(redis); err != nil {
			log.Errorln(err)
			return nil
		}
	}
	return nil
}

func (c *Controller) setOwnerReferenceToOffshoots(redis *api.Redis) error {
	owner := metav1.NewControllerRef(redis, api.SchemeGroupVersion.WithKind(api.ResourceKindRedis))
	selector := labels.SelectorFromSet(redis.OffshootSelectors())

	// delete PVC for both "wipeOut" and "delete" TerminationPolicy.
	return dynamic_util.EnsureOwnerReferenceForSelector(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		redis.Namespace,
		selector,
		owner)
}

func (c *Controller) removeOwnerReferenceFromOffshoots(redis *api.Redis) error {
	// First, Get LabelSelector for Other Components
	labelSelector := labels.SelectorFromSet(redis.OffshootSelectors())

	return dynamic_util.RemoveOwnerReferenceForSelector(
		context.TODO(),
		c.DynamicClient,
		core.SchemeGroupVersion.WithResource("persistentvolumeclaims"),
		redis.Namespace,
		labelSelector,
		redis)
}
