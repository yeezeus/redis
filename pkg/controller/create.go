package controller

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	kutildb "github.com/k8sdb/apimachinery/client/typed/kubedb/v1alpha1/util"
	"github.com/k8sdb/apimachinery/pkg/docker"
	"github.com/k8sdb/apimachinery/pkg/eventer"
	"github.com/k8sdb/apimachinery/pkg/storage"
	apps "k8s.io/api/apps/v1beta1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// Duration in Minute
	// Check whether pod under StatefulSet is running or not
	// Continue checking for this duration until failure
	durationCheckStatefulSet = time.Minute * 30
)

func (c *Controller) findService(redis *tapi.Xdb) (bool, error) {
	name := redis.OffshootName()
	service, err := c.Client.CoreV1().Services(redis.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if service.Spec.Selector[tapi.LabelDatabaseName] != name {
		return false, fmt.Errorf(`Intended service "%v" already exists`, name)
	}

	return true, nil
}

func (c *Controller) createService(redis *tapi.Xdb) error {
	svc := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   redis.OffshootName(),
			Labels: redis.OffshootLabels(),
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
			// TODO: Use appropriate port for your service
			},
			Selector: redis.OffshootLabels(),
		},
	}
	if redis.Spec.Monitor != nil &&
		redis.Spec.Monitor.Agent == tapi.AgentCoreosPrometheus &&
		redis.Spec.Monitor.Prometheus != nil {
		svc.Spec.Ports = append(svc.Spec.Ports, core.ServicePort{
			Name:       tapi.PrometheusExporterPortName,
			Port:       tapi.PrometheusExporterPortNumber,
			TargetPort: intstr.FromString(tapi.PrometheusExporterPortName),
		})
	}

	if _, err := c.Client.CoreV1().Services(redis.Namespace).Create(svc); err != nil {
		return err
	}

	return nil
}

func (c *Controller) findStatefulSet(redis *tapi.Xdb) (bool, error) {
	// SatatefulSet for Xdb database
	statefulSet, err := c.Client.AppsV1beta1().StatefulSets(redis.Namespace).Get(redis.OffshootName(), metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}

	if statefulSet.Labels[tapi.LabelDatabaseKind] != tapi.ResourceKindXdb {
		return false, fmt.Errorf(`Intended statefulSet "%v" already exists`, redis.OffshootName())
	}

	return true, nil
}

func (c *Controller) createStatefulSet(redis *tapi.Xdb) (*apps.StatefulSet, error) {
	// SatatefulSet for Xdb database
	statefulSet := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        redis.OffshootName(),
			Namespace:   redis.Namespace,
			Labels:      redis.StatefulSetLabels(),
			Annotations: redis.StatefulSetAnnotations(),
		},
		Spec: apps.StatefulSetSpec{
			Replicas:    types.Int32P(1),
			ServiceName: c.opt.GoverningService,
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: redis.OffshootLabels(),
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name: tapi.ResourceNameXdb,
							//TODO: Use correct image. Its a template
							Image:           fmt.Sprintf("%s:%s", docker.ImageXdb, redis.Spec.Version),
							ImagePullPolicy: core.PullIfNotPresent,
							Ports:           []core.ContainerPort{
							//TODO: Use appropriate port for your container
							},
							Resources: redis.Spec.Resources,
							VolumeMounts: []core.VolumeMount{
								//TODO: Add Secret volume if necessary
								{
									Name:      "data",
									MountPath: "/var/pv",
								},
							},
							Args: []string{ /*TODO Add args if necessary*/ },
						},
					},
					NodeSelector:  redis.Spec.NodeSelector,
					Affinity:      redis.Spec.Affinity,
					SchedulerName: redis.Spec.SchedulerName,
					Tolerations:   redis.Spec.Tolerations,
				},
			},
		},
	}

	if redis.Spec.Monitor != nil &&
		redis.Spec.Monitor.Agent == tapi.AgentCoreosPrometheus &&
		redis.Spec.Monitor.Prometheus != nil {
		exporter := core.Container{
			Name: "exporter",
			Args: []string{
				"export",
				fmt.Sprintf("--address=:%d", tapi.PrometheusExporterPortNumber),
				"--v=3",
			},
			Image:           docker.ImageOperator + ":" + c.opt.ExporterTag,
			ImagePullPolicy: core.PullIfNotPresent,
			Ports: []core.ContainerPort{
				{
					Name:          tapi.PrometheusExporterPortName,
					Protocol:      core.ProtocolTCP,
					ContainerPort: int32(tapi.PrometheusExporterPortNumber),
				},
			},
		}
		statefulSet.Spec.Template.Spec.Containers = append(statefulSet.Spec.Template.Spec.Containers, exporter)
	}

	// ---> Start
	//TODO: Use following if secret is necessary
	// otherwise remove
	if redis.Spec.DatabaseSecret == nil {
		secretVolumeSource, err := c.createDatabaseSecret(redis)
		if err != nil {
			return nil, err
		}

		_redis, err := kutildb.TryPatchXdb(c.ExtClient, redis.ObjectMeta, func(in *tapi.Xdb) *tapi.Xdb {
			in.Spec.DatabaseSecret = secretVolumeSource
			return in
		})
		if err != nil {
			c.recorder.Eventf(redis.ObjectReference(), core.EventTypeWarning, eventer.EventReasonFailedToUpdate, err.Error())
			return nil, err
		}
		redis = _redis
	}

	// Add secretVolume for authentication
	addSecretVolume(statefulSet, redis.Spec.DatabaseSecret)
	// --- > End

	// Add Data volume for StatefulSet
	addDataVolume(statefulSet, redis.Spec.Storage)

	// ---> Start
	//TODO: Use following if supported
	// otherwise remove

	// Add InitialScript to run at startup
	if redis.Spec.Init != nil && redis.Spec.Init.ScriptSource != nil {
		addInitialScript(statefulSet, redis.Spec.Init.ScriptSource)
	}
	// ---> End

	if c.opt.EnableRbac {
		// Ensure ClusterRoles for database statefulsets
		if err := c.createRBACStuff(redis); err != nil {
			return nil, err
		}

		statefulSet.Spec.Template.Spec.ServiceAccountName = redis.Name
	}

	if _, err := c.Client.AppsV1beta1().StatefulSets(statefulSet.Namespace).Create(statefulSet); err != nil {
		return nil, err
	}

	return statefulSet, nil
}

func (c *Controller) findSecret(secretName, namespace string) (bool, error) {
	secret, err := c.Client.CoreV1().Secrets(namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	if secret == nil {
		return false, nil
	}

	return true, nil
}

// ---> start
//TODO: Use this method to create secret dynamically
// otherwise remove this method
func (c *Controller) createDatabaseSecret(redis *tapi.Xdb) (*core.SecretVolumeSource, error) {
	authSecretName := redis.Name + "-admin-auth"

	found, err := c.findSecret(authSecretName, redis.Namespace)
	if err != nil {
		return nil, err
	}

	if !found {

		secret := &core.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: authSecretName,
				Labels: map[string]string{
					tapi.LabelDatabaseKind: tapi.ResourceKindXdb,
				},
			},
			Type: core.SecretTypeOpaque,
			Data: make(map[string][]byte), // Add secret data
		}
		if _, err := c.Client.CoreV1().Secrets(redis.Namespace).Create(secret); err != nil {
			return nil, err
		}
	}

	return &core.SecretVolumeSource{
		SecretName: authSecretName,
	}, nil
}

// ---> End

// ---> Start
//TODO: Use this method to add secret volume
// otherwise remove this method
func addSecretVolume(statefulSet *apps.StatefulSet, secretVolume *core.SecretVolumeSource) error {
	statefulSet.Spec.Template.Spec.Volumes = append(statefulSet.Spec.Template.Spec.Volumes,
		core.Volume{
			Name: "secret",
			VolumeSource: core.VolumeSource{
				Secret: secretVolume,
			},
		},
	)
	return nil
}

// ---> End

func addDataVolume(statefulSet *apps.StatefulSet, pvcSpec *core.PersistentVolumeClaimSpec) {
	if pvcSpec != nil {
		if len(pvcSpec.AccessModes) == 0 {
			pvcSpec.AccessModes = []core.PersistentVolumeAccessMode{
				core.ReadWriteOnce,
			}
			log.Infof(`Using "%v" as AccessModes in "%v"`, core.ReadWriteOnce, *pvcSpec)
		}
		// volume claim templates
		// Dynamically attach volume
		statefulSet.Spec.VolumeClaimTemplates = []core.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "data",
					Annotations: map[string]string{
						"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
					},
				},
				Spec: *pvcSpec,
			},
		}
	} else {
		// Attach Empty directory
		statefulSet.Spec.Template.Spec.Volumes = append(
			statefulSet.Spec.Template.Spec.Volumes,
			core.Volume{
				Name: "data",
				VolumeSource: core.VolumeSource{
					EmptyDir: &core.EmptyDirVolumeSource{},
				},
			},
		)
	}
}

// ---> Start
//TODO: Use this method to add initial script, if supported
// Otherwise, remove it
func addInitialScript(statefulSet *apps.StatefulSet, script *tapi.ScriptSourceSpec) {
	statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts = append(statefulSet.Spec.Template.Spec.Containers[0].VolumeMounts,
		core.VolumeMount{
			Name:      "initial-script",
			MountPath: "/var/db-script",
		},
	)
	statefulSet.Spec.Template.Spec.Containers[0].Args = []string{
		// Add additional args
		script.ScriptPath,
	}

	statefulSet.Spec.Template.Spec.Volumes = append(statefulSet.Spec.Template.Spec.Volumes,
		core.Volume{
			Name:         "initial-script",
			VolumeSource: script.VolumeSource,
		},
	)
}

// ---> End

func (c *Controller) createDormantDatabase(redis *tapi.Xdb) (*tapi.DormantDatabase, error) {
	dormantDb := &tapi.DormantDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      redis.Name,
			Namespace: redis.Namespace,
			Labels: map[string]string{
				tapi.LabelDatabaseKind: tapi.ResourceKindXdb,
			},
		},
		Spec: tapi.DormantDatabaseSpec{
			Origin: tapi.Origin{
				ObjectMeta: metav1.ObjectMeta{
					Name:        redis.Name,
					Namespace:   redis.Namespace,
					Labels:      redis.Labels,
					Annotations: redis.Annotations,
				},
				Spec: tapi.OriginSpec{
					Xdb: &redis.Spec,
				},
			},
		},
	}

	initSpec, _ := json.Marshal(redis.Spec.Init)
	if initSpec != nil {
		dormantDb.Annotations = map[string]string{
			tapi.XdbInitSpec: string(initSpec),
		}
	}

	dormantDb.Spec.Origin.Spec.Xdb.Init = nil

	return c.ExtClient.DormantDatabases(dormantDb.Namespace).Create(dormantDb)
}

func (c *Controller) reCreateXdb(redis *tapi.Xdb) error {
	_redis := &tapi.Xdb{
		ObjectMeta: metav1.ObjectMeta{
			Name:        redis.Name,
			Namespace:   redis.Namespace,
			Labels:      redis.Labels,
			Annotations: redis.Annotations,
		},
		Spec:   redis.Spec,
		Status: redis.Status,
	}

	if _, err := c.ExtClient.Xdbs(_redis.Namespace).Create(_redis); err != nil {
		return err
	}

	return nil
}

const (
	SnapshotProcess_Restore  = "restore"
	snapshotType_DumpRestore = "dump-restore"
)

func (c *Controller) createRestoreJob(redis *tapi.Xdb, snapshot *tapi.Snapshot) (*batch.Job, error) {
	databaseName := redis.Name
	jobName := snapshot.OffshootName()
	jobLabel := map[string]string{
		tapi.LabelDatabaseName: databaseName,
		tapi.LabelJobType:      SnapshotProcess_Restore,
	}
	backupSpec := snapshot.Spec.SnapshotStorageSpec
	bucket, err := backupSpec.Container()
	if err != nil {
		return nil, err
	}

	// Get PersistentVolume object for Backup Util pod.
	persistentVolume, err := c.getVolumeForSnapshot(redis.Spec.Storage, jobName, redis.Namespace)
	if err != nil {
		return nil, err
	}

	// Folder name inside Cloud bucket where backup will be uploaded
	folderName, _ := snapshot.Location()

	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:   jobName,
			Labels: jobLabel,
		},
		Spec: batch.JobSpec{
			Template: core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: jobLabel,
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name: SnapshotProcess_Restore,
							//TODO: Use appropriate image
							Image: fmt.Sprintf("%s:%s", docker.ImageXdb, redis.Spec.Version),
							Args: []string{
								fmt.Sprintf(`--process=%s`, SnapshotProcess_Restore),
								fmt.Sprintf(`--host=%s`, databaseName),
								fmt.Sprintf(`--bucket=%s`, bucket),
								fmt.Sprintf(`--folder=%s`, folderName),
								fmt.Sprintf(`--snapshot=%s`, snapshot.Name),
							},
							Resources: snapshot.Spec.Resources,
							VolumeMounts: []core.VolumeMount{
								//TODO: Mount secret volume if necessary
								{
									Name:      persistentVolume.Name,
									MountPath: "/var/" + snapshotType_DumpRestore + "/",
								},
								{
									Name:      "osmconfig",
									MountPath: storage.SecretMountPath,
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []core.Volume{
						//TODO: Add secret volume if necessary
						// Check postgres repository for example
						{
							Name:         persistentVolume.Name,
							VolumeSource: persistentVolume.VolumeSource,
						},
						{
							Name: "osmconfig",
							VolumeSource: core.VolumeSource{
								Secret: &core.SecretVolumeSource{
									SecretName: snapshot.Name,
								},
							},
						},
					},
					RestartPolicy: core.RestartPolicyNever,
				},
			},
		},
	}
	if snapshot.Spec.SnapshotStorageSpec.Local != nil {
		job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, core.VolumeMount{
			Name:      "local",
			MountPath: snapshot.Spec.SnapshotStorageSpec.Local.Path,
		})
		volume := core.Volume{
			Name:         "local",
			VolumeSource: snapshot.Spec.SnapshotStorageSpec.Local.VolumeSource,
		}
		job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, volume)
	}
	return c.Client.BatchV1().Jobs(redis.Namespace).Create(job)
}
