package framework

import (
	"bytes"
	"fmt"
	"strings"

	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteInForeground() *metav1.DeleteOptions {
	policy := metav1.DeletePropagationForeground
	return &metav1.DeleteOptions{PropagationPolicy: &policy}
}

func (fi *Invocation) GetPod(meta metav1.ObjectMeta) (*core.Pod, error) {
	podList, err := fi.kubeClient.CoreV1().Pods(meta.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, pod := range podList.Items {
		if bytes.HasPrefix([]byte(pod.Name), []byte(meta.Name)) {
			return &pod, nil
		}
	}
	return nil, fmt.Errorf("no pod found for workload %v", meta.Name)
}

func (fi *Invocation) GetCustomConfig(configs []string) *core.ConfigMap {
	return &core.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fi.app,
			Namespace: fi.namespace,
		},
		Data: map[string]string{
			"redis.conf": strings.Join(configs, "\n"),
		},
	}
}

func (fi *Invocation) CreateConfigMap(obj *core.ConfigMap) error {
	_, err := fi.kubeClient.CoreV1().ConfigMaps(obj.Namespace).Create(obj)
	return err
}

func (fi *Invocation) DeleteConfigMap(meta metav1.ObjectMeta) error {
	err := fi.kubeClient.CoreV1().ConfigMaps(meta.Namespace).Delete(meta.Name, deleteInForeground())
	if err != nil && !kerr.IsNotFound(err) {
		return err
	}
	return nil
}
