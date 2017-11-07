package validator

import (
	"fmt"

	tapi "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/k8sdb/apimachinery/pkg/docker"
	amv "github.com/k8sdb/apimachinery/pkg/validator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// TODO: Change method name. ValidateXdb -> Validate<--->
func ValidateXdb(client kubernetes.Interface, redis *tapi.Xdb) error {
	if redis.Spec.Version == "" {
		return fmt.Errorf(`Object 'Version' is missing in '%v'`, redis.Spec)
	}

	// Set Database Image version
	version := redis.Spec.Version
	// TODO: docker.ImageXdb should hold correct image name
	if err := docker.CheckDockerImageVersion(docker.ImageXdb, version); err != nil {
		return fmt.Errorf(`Image %v:%v not found`, docker.ImageXdb, version)
	}

	if redis.Spec.Storage != nil {
		var err error
		if err = amv.ValidateStorage(client, redis.Spec.Storage); err != nil {
			return err
		}
	}

	// ---> Start
	// TODO: Use following if database needs/supports authentication secret
	// otherwise, delete
	databaseSecret := redis.Spec.DatabaseSecret
	if databaseSecret != nil {
		if _, err := client.CoreV1().Secrets(redis.Namespace).Get(databaseSecret.SecretName, metav1.GetOptions{}); err != nil {
			return err
		}
	}
	// ---> End

	backupScheduleSpec := redis.Spec.BackupSchedule
	if backupScheduleSpec != nil {
		if err := amv.ValidateBackupSchedule(client, backupScheduleSpec, redis.Namespace); err != nil {
			return err
		}
	}

	monitorSpec := redis.Spec.Monitor
	if monitorSpec != nil {
		if err := amv.ValidateMonitorSpec(monitorSpec); err != nil {
			return err
		}

	}
	return nil
}
