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
package framework

import (
	"context"
	"fmt"
	"strconv"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/types"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	meta_util "kmodules.xyz/client-go/meta"
)

const (
	kindEviction = "Eviction"
)

func (fi *Invocation) Redis() *api.Redis {
	return &api.Redis{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("redis"),
			Namespace: fi.namespace,
			Labels: map[string]string{
				"app": fi.app,
			},
		},
		Spec: api.RedisSpec{
			Version: DBCatalogName,
			UpdateStrategy: apps.StatefulSetUpdateStrategy{
				Type: apps.RollingUpdateStatefulSetStrategyType,
			},
			TerminationPolicy: api.TerminationPolicyHalt,
			Storage: &core.PersistentVolumeClaimSpec{
				Resources: core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceStorage: resource.MustParse("1Gi"),
					},
				},
				StorageClassName: types.StringP(fi.StorageClass),
			},
		},
	}
}

func (fi *Invocation) RedisCluster() *api.Redis {
	redis := fi.Redis()
	redis.Spec.Mode = api.RedisModeCluster
	redis.Spec.Cluster = &api.RedisClusterSpec{
		Master:   types.Int32P(3),
		Replicas: types.Int32P(1),
	}

	return redis
}

func (f *Framework) CreateRedis(obj *api.Redis) error {
	_, err := f.dbClient.KubedbV1alpha1().Redises(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
	return err
}

func (f *Framework) GetRedis(meta metav1.ObjectMeta) (*api.Redis, error) {
	return f.dbClient.KubedbV1alpha1().Redises(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
}

func (f *Framework) PatchRedis(meta metav1.ObjectMeta, transform func(*api.Redis) *api.Redis) (*api.Redis, error) {
	redis, err := f.dbClient.KubedbV1alpha1().Redises(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	redis, _, err = util.PatchRedis(context.TODO(), f.dbClient.KubedbV1alpha1(), redis, transform, metav1.PatchOptions{})
	return redis, err
}

func (f *Framework) DeleteRedis(meta metav1.ObjectMeta) error {
	return f.dbClient.KubedbV1alpha1().Redises(meta.Namespace).Delete(context.TODO(), meta.Name, deleteInForeground())
}

func (f *Framework) EventuallyRedis(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			_, err := f.dbClient.KubedbV1alpha1().Redises(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
			if err != nil {
				if kerr.IsNotFound(err) {
					return false
				}
				Expect(err).NotTo(HaveOccurred())
			}
			return true
		},
		time.Minute*12,
		time.Second*5,
	)
}

func (f *Framework) EventuallyRedisPhase(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() api.DatabasePhase {
			db, err := f.dbClient.KubedbV1alpha1().Redises(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return db.Status.Phase
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) EventuallyRedisRunning(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			redis, err := f.dbClient.KubedbV1alpha1().Redises(meta.Namespace).Get(context.TODO(), meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return redis.Status.Phase == api.DatabasePhaseRunning
		},
		time.Minute*13,
		time.Second*5,
	)
}

func (f *Framework) CleanRedis() {
	redisList, err := f.dbClient.KubedbV1alpha1().Redises(f.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return
	}
	for _, e := range redisList.Items {
		if _, _, err := util.PatchRedis(context.TODO(), f.dbClient.KubedbV1alpha1(), &e, func(in *api.Redis) *api.Redis {
			in.ObjectMeta.Finalizers = nil
			in.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
			return in
		}, metav1.PatchOptions{}); err != nil {
			fmt.Printf("error Patching Redis. error: %v", err)
		}
	}
	if err := f.dbClient.KubedbV1alpha1().Redises(f.namespace).DeleteCollection(context.TODO(), deleteInForeground(), metav1.ListOptions{}); err != nil {
		fmt.Printf("error in deletion of Redis. Error: %v", err)
	}
}

func (f *Framework) EvictPodsFromStatefulSet(meta metav1.ObjectMeta) error {
	var err error
	labelSelector := labels.Set{
		meta_util.ManagedByLabelKey: api.GenericKey,
		api.LabelDatabaseKind:       api.ResourceKindRedis,
		api.LabelDatabaseName:       meta.GetName(),
	}

	// get sts in the namespace
	stsList, err := f.kubeClient.AppsV1().StatefulSets(meta.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: labelSelector.String()})
	if err != nil {
		return err
	}

	if len(stsList.Items) < 1 {
		return fmt.Errorf("found no statefulset in namespace %s with specific labels", meta.Namespace)
	}

	for _, sts := range stsList.Items {
		// if PDB is not found, send error
		var pdb *policy.PodDisruptionBudget
		pdb, err = f.kubeClient.PolicyV1beta1().PodDisruptionBudgets(sts.Namespace).Get(context.TODO(), sts.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		eviction := &policy.Eviction{
			TypeMeta: metav1.TypeMeta{
				APIVersion: policy.SchemeGroupVersion.String(),
				Kind:       kindEviction,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      sts.Name,
				Namespace: sts.Namespace,
			},
			DeleteOptions: &metav1.DeleteOptions{},
		}

		if pdb.Spec.MaxUnavailable == nil {
			return fmt.Errorf("found pdb %s spec.maxUnavailable nil", pdb.Name)
		}

		// try to evict as many pod as allowed in pdb. No err should occur
		maxUnavailable := pdb.Spec.MaxUnavailable.IntValue()
		for i := 0; i < maxUnavailable; i++ {
			eviction.Name = sts.Name + "-" + strconv.Itoa(i)

			err := f.kubeClient.PolicyV1beta1().Evictions(eviction.Namespace).Evict(context.TODO(), eviction)
			if err != nil {
				return err
			}
		}

		// try to evict one extra pod. TooManyRequests err should occur
		eviction.Name = sts.Name + "-" + strconv.Itoa(maxUnavailable-1)

		err = f.kubeClient.PolicyV1beta1().Evictions(eviction.Namespace).Evict(context.TODO(), eviction)
		if kerr.IsTooManyRequests(err) {
			err = nil
		} else if err != nil {
			return err
		} else {
			return fmt.Errorf("expected pod %s/%s to be not evicted due to pdb %s", sts.Namespace, eviction.Name, pdb.Name)
		}
	}
	return err
}
