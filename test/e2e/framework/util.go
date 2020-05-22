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
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	shell "github.com/codeskyblue/go-sh"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func deleteInForeground() metav1.DeleteOptions {
	policy := metav1.DeletePropagationForeground
	return metav1.DeleteOptions{PropagationPolicy: &policy}
}

func (fi *Invocation) GetPod(meta metav1.ObjectMeta) (*core.Pod, error) {
	podList, err := fi.kubeClient.CoreV1().Pods(meta.Namespace).List(context.TODO(), metav1.ListOptions{})
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
	_, err := fi.kubeClient.CoreV1().ConfigMaps(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
	return err
}

func (fi *Invocation) DeleteConfigMap(meta metav1.ObjectMeta) error {
	err := fi.kubeClient.CoreV1().ConfigMaps(meta.Namespace).Delete(context.TODO(), meta.Name, deleteInForeground())
	if err != nil && !kerr.IsNotFound(err) {
		return err
	}
	return nil
}

func (f *Framework) EventuallyWipedOut(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() error {
			labelMap := map[string]string{
				api.LabelDatabaseName: meta.Name,
				api.LabelDatabaseKind: api.ResourceKindRedis,
			}
			labelSelector := labels.SelectorFromSet(labelMap)

			// check if pvcs is wiped out
			pvcList, err := f.kubeClient.CoreV1().PersistentVolumeClaims(meta.Namespace).List(
				context.TODO(),
				metav1.ListOptions{
					LabelSelector: labelSelector.String(),
				},
			)
			if err != nil {
				return err
			}
			if len(pvcList.Items) > 0 {
				return fmt.Errorf("PVCs have not wiped out yet")
			}

			// check if secrets are wiped out
			secretList, err := f.kubeClient.CoreV1().Secrets(meta.Namespace).List(
				context.TODO(),
				metav1.ListOptions{
					LabelSelector: labelSelector.String(),
				},
			)
			if err != nil {
				return err
			}
			if len(secretList.Items) > 0 {
				return fmt.Errorf("secrets have not wiped out yet")
			}

			// check if appbinds are wiped out
			appBindingList, err := f.appCatalogClient.AppBindings(meta.Namespace).List(
				context.TODO(),
				metav1.ListOptions{
					LabelSelector: labelSelector.String(),
				},
			)
			if err != nil {
				return err
			}
			if len(appBindingList.Items) > 0 {
				return fmt.Errorf("appBindings have not wiped out yet")
			}

			return nil
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) PrintDebugHelpers() {
	sh := shell.NewSession()

	fmt.Println("\n======================================[ Describe Pod ]===================================================")
	if err := sh.Command("/usr/bin/kubectl", "describe", "po", "-n", f.Namespace()).Run(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("\n======================================[ Describe Redis ]===================================================")
	if err := sh.Command("/usr/bin/kubectl", "describe", "rd", "-n", f.Namespace()).Run(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("\n======================================[ Describe Nodes ]===================================================")
	if err := sh.Command("/usr/bin/kubectl", "describe", "nodes").Run(); err != nil {
		fmt.Println(err)
	}
}
