package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/appscode/log"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	kutildb "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	"github.com/k8sdb/apimachinery/pkg/storage"
	"github.com/k8sdb/redis/pkg/validator"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Use your resource instead of *tapi.Xdb
func (c *Controller) create(redis *tapi.Xdb) error {
	// TODO: Use correct TryPatch method
	_, err := kutildb.TryPatchXdb(c.ExtClient, redis.ObjectMeta, func(in *tapi.Xdb) *tapi.Xdb {
		t := metav1.Now()
		in.Status.CreationTime = &t
		in.Status.Phase = tapi.DatabasePhaseCreating
		return in
	})

	if err != nil {
		c.recorder.Eventf(redis.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}

	if err := validator.ValidateXdb(c.Client, redis); err != nil {
		c.recorder.Event(redis.ObjectReference(), core.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}
	// Event for successful validation
	c.recorder.Event(
		redis.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulValidate,
		"Successfully validate Xdb",
	)

	// Check DormantDatabase
	matched, err := c.matchDormantDatabase(redis)
	if err != nil {
		return err
	}
	if matched {
		//TODO: Use Annotation Key
		redis.Annotations = map[string]string{
			"kubedb.com/ignore": "",
		}
		if err := c.ExtClient.Xdbs(redis.Namespace).Delete(redis.Name, &metav1.DeleteOptions{}); err != nil {
			return fmt.Errorf(
				`Failed to resume Xdb "%v" from DormantDatabase "%v". Error: %v`,
				redis.Name,
				redis.Name,
				err,
			)
		}

		_, err := kutildb.TryPatchDormantDatabase(c.ExtClient, redis.ObjectMeta, func(in *tapi.DormantDatabase) *tapi.DormantDatabase {
			in.Spec.Resume = true
			return in
		})
		if err != nil {
			c.recorder.Eventf(redis.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}

		return nil
	}

	// Event for notification that kubernetes objects are creating
	c.recorder.Event(redis.ObjectReference(), core.EventTypeNormal, eventer.EventReasonCreating, "Creating Kubernetes objects")

	// create Governing Service
	governingService := c.opt.GoverningService
	if err := c.CreateGoverningService(governingService, redis.Namespace); err != nil {
		c.recorder.Eventf(
			redis.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create Service: "%v". Reason: %v`,
			governingService,
			err,
		)
		return err
	}

	// ensure database Service
	if err := c.ensureService(redis); err != nil {
		return err
	}

	// ensure database StatefulSet
	if err := c.ensureStatefulSet(redis); err != nil {
		return err
	}

	c.recorder.Event(
		redis.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		"Successfully created Xdb",
	)

	// Ensure Schedule backup
	c.ensureBackupScheduler(redis)

	if redis.Spec.Monitor != nil {
		if err := c.addMonitor(redis); err != nil {
			c.recorder.Eventf(
				redis.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				"Failed to add monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			redis.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulCreate,
			"Successfully added monitoring system.",
		)
	}
	return nil
}

func (c *Controller) matchDormantDatabase(redis *tapi.Xdb) (bool, error) {
	// Check if DormantDatabase exists or not
	dormantDb, err := c.ExtClient.DormantDatabases(redis.Namespace).Get(redis.Name, metav1.GetOptions{})
	if err != nil {
		if !kerr.IsNotFound(err) {
			c.recorder.Eventf(
				redis.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToGet,
				`Fail to get DormantDatabase: "%v". Reason: %v`,
				redis.Name,
				err,
			)
			return false, err
		}
		return false, nil
	}

	var sendEvent = func(message string) (bool, error) {
		c.recorder.Event(
			redis.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			message,
		)
		return false, errors.New(message)
	}

	// Check DatabaseKind
	// TODO: Change tapi.ResourceKindXdb
	if dormantDb.Labels[tapi.LabelDatabaseKind] != tapi.ResourceKindXdb {
		return sendEvent(fmt.Sprintf(`Invalid Xdb: "%v". Exists DormantDatabase "%v" of different Kind`,
			redis.Name, dormantDb.Name))
	}

	// Check InitSpec
	// TODO: Change tapi.XdbInitSpec
	initSpecAnnotationStr := dormantDb.Annotations[tapi.XdbInitSpec]
	if initSpecAnnotationStr != "" {
		var initSpecAnnotation *tapi.InitSpec
		if err := json.Unmarshal([]byte(initSpecAnnotationStr), &initSpecAnnotation); err != nil {
			return sendEvent(err.Error())
		}

		if redis.Spec.Init != nil {
			if !reflect.DeepEqual(initSpecAnnotation, redis.Spec.Init) {
				return sendEvent("InitSpec mismatches with DormantDatabase annotation")
			}
		}
	}

	// Check Origin Spec
	drmnOriginSpec := dormantDb.Spec.Origin.Spec.Xdb
	originalSpec := redis.Spec
	originalSpec.Init = nil

	// ---> Start
	// TODO: Use following part if database secret is supported
	// Otherwise, remove it
	if originalSpec.DatabaseSecret == nil {
		originalSpec.DatabaseSecret = &core.SecretVolumeSource{
			SecretName: redis.Name + "-admin-auth",
		}
	}
	// ---> End

	if !reflect.DeepEqual(drmnOriginSpec, &originalSpec) {
		return sendEvent("Xdb spec mismatches with OriginSpec in DormantDatabases")
	}

	return true, nil
}

func (c *Controller) ensureService(redis *tapi.Xdb) error {
	// Check if service name exists
	found, err := c.findService(redis)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// create database Service
	if err := c.createService(redis); err != nil {
		c.recorder.Eventf(
			redis.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to create Service. Reason: %v",
			err,
		)
		return err
	}
	return nil
}

func (c *Controller) ensureStatefulSet(redis *tapi.Xdb) error {
	found, err := c.findStatefulSet(redis)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	// Create statefulSet for Xdb database
	statefulSet, err := c.createStatefulSet(redis)
	if err != nil {
		c.recorder.Eventf(
			redis.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			"Failed to create StatefulSet. Reason: %v",
			err,
		)
		return err
	}

	// Check StatefulSet Pod status
	if err := c.CheckStatefulSetPodStatus(statefulSet, durationCheckStatefulSet); err != nil {
		c.recorder.Eventf(
			redis.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToStart,
			`Failed to create StatefulSet. Reason: %v`,
			err,
		)
		return err
	} else {
		c.recorder.Event(
			redis.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulCreate,
			"Successfully created StatefulSet",
		)
	}

	if redis.Spec.Init != nil && redis.Spec.Init.SnapshotSource != nil {
		// TODO: Use correct TryPatch method
		_, err := kutildb.TryPatchXdb(c.ExtClient, redis.ObjectMeta, func(in *tapi.Xdb) *tapi.Xdb {
			in.Status.Phase = tapi.DatabasePhaseInitializing
			return in
		})
		if err != nil {
			c.recorder.Eventf(redis, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return err
		}

		if err := c.initialize(redis); err != nil {
			c.recorder.Eventf(
				redis.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToInitialize,
				"Failed to initialize. Reason: %v",
				err,
			)
		}
	}

	// TODO: Use correct TryPatch method
	_, err = kutildb.TryPatchXdb(c.ExtClient, redis.ObjectMeta, func(in *tapi.Xdb) *tapi.Xdb {
		in.Status.Phase = tapi.DatabasePhaseRunning
		return in
	})
	if err != nil {
		c.recorder.Eventf(redis, core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
		return err
	}
	return nil
}

func (c *Controller) ensureBackupScheduler(redis *tapi.Xdb) {
	// Setup Schedule backup
	if redis.Spec.BackupSchedule != nil {
		err := c.cronController.ScheduleBackup(redis, redis.ObjectMeta, redis.Spec.BackupSchedule)
		if err != nil {
			c.recorder.Eventf(
				redis.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToSchedule,
				"Failed to schedule snapshot. Reason: %v",
				err,
			)
			log.Errorln(err)
		}
	} else {
		c.cronController.StopBackupScheduling(redis.ObjectMeta)
	}
}

const (
	durationCheckRestoreJob = time.Minute * 30
)

func (c *Controller) initialize(redis *tapi.Xdb) error {
	snapshotSource := redis.Spec.Init.SnapshotSource
	// Event for notification that kubernetes objects are creating
	c.recorder.Eventf(
		redis.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonInitializing,
		`Initializing from Snapshot: "%v"`,
		snapshotSource.Name,
	)

	namespace := snapshotSource.Namespace
	if namespace == "" {
		namespace = redis.Namespace
	}
	snapshot, err := c.ExtClient.Snapshots(namespace).Get(snapshotSource.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	secret, err := storage.NewOSMSecret(c.Client, snapshot)
	if err != nil {
		return err
	}
	_, err = c.Client.CoreV1().Secrets(secret.Namespace).Create(secret)
	if err != nil {
		return err
	}

	job, err := c.createRestoreJob(redis, snapshot)
	if err != nil {
		return err
	}

	jobSuccess := c.CheckDatabaseRestoreJob(job, redis, c.recorder, durationCheckRestoreJob)
	if jobSuccess {
		c.recorder.Event(
			redis.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulInitialize,
			"Successfully completed initialization",
		)
	} else {
		c.recorder.Event(
			redis.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToInitialize,
			"Failed to complete initialization",
		)
	}
	return nil
}

func (c *Controller) pause(redis *tapi.Xdb) error {
	if redis.Annotations != nil {
		if val, found := redis.Annotations["kubedb.com/ignore"]; found {
			//TODO: Add Event Reason "Ignored"
			c.recorder.Event(redis.ObjectReference(), core.EventTypeNormal, "Ignored", val)
			return nil
		}
	}

	c.recorder.Event(redis.ObjectReference(), core.EventTypeNormal, eventer.EventReasonPausing, "Pausing Xdb")

	if redis.Spec.DoNotPause {
		c.recorder.Eventf(
			redis.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToPause,
			`Xdb "%v" is locked.`,
			redis.Name,
		)

		if err := c.reCreateXdb(redis); err != nil {
			c.recorder.Eventf(
				redis.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToCreate,
				`Failed to recreate Xdb: "%v". Reason: %v`,
				redis.Name,
				err,
			)
			return err
		}
		return nil
	}

	if _, err := c.createDormantDatabase(redis); err != nil {
		c.recorder.Eventf(
			redis.ObjectReference(),
			core.EventTypeWarning,
			eventer.EventReasonFailedToCreate,
			`Failed to create DormantDatabase: "%v". Reason: %v`,
			redis.Name,
			err,
		)
		return err
	}
	c.recorder.Eventf(
		redis.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulCreate,
		`Successfully created DormantDatabase: "%v"`,
		redis.Name,
	)

	c.cronController.StopBackupScheduling(redis.ObjectMeta)

	if redis.Spec.Monitor != nil {
		if err := c.deleteMonitor(redis); err != nil {
			c.recorder.Eventf(
				redis.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToDelete,
				"Failed to delete monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			redis.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorDelete,
			"Successfully deleted monitoring system.",
		)
	}
	return nil
}

func (c *Controller) update(oldXdb, updatedXdb *tapi.Xdb) error {
	if err := validator.ValidateXdb(c.Client, updatedXdb); err != nil {
		c.recorder.Event(updatedXdb.ObjectReference(), core.EventTypeWarning, eventer.EventReasonInvalid, err.Error())
		return err
	}
	// Event for successful validation
	c.recorder.Event(
		updatedXdb.ObjectReference(),
		core.EventTypeNormal,
		eventer.EventReasonSuccessfulValidate,
		"Successfully validate Xdb",
	)

	if err := c.ensureService(updatedXdb); err != nil {
		return err
	}
	if err := c.ensureStatefulSet(updatedXdb); err != nil {
		return err
	}

	if !reflect.DeepEqual(updatedXdb.Spec.BackupSchedule, oldXdb.Spec.BackupSchedule) {
		c.ensureBackupScheduler(updatedXdb)
	}

	if !reflect.DeepEqual(oldXdb.Spec.Monitor, updatedXdb.Spec.Monitor) {
		if err := c.updateMonitor(oldXdb, updatedXdb); err != nil {
			c.recorder.Eventf(
				updatedXdb.ObjectReference(),
				core.EventTypeWarning,
				eventer.EventReasonFailedToUpdate,
				"Failed to update monitoring system. Reason: %v",
				err,
			)
			log.Errorln(err)
			return nil
		}
		c.recorder.Event(
			updatedXdb.ObjectReference(),
			core.EventTypeNormal,
			eventer.EventReasonSuccessfulMonitorUpdate,
			"Successfully updated monitoring system.",
		)

	}
	return nil
}
