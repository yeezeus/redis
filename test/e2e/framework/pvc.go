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
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (f *Framework) EventuallyPVCCount(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	labelMap := map[string]string{
		api.LabelDatabaseKind: api.ResourceKindMemcached,
		api.LabelDatabaseName: meta.Name,
	}
	labelSelector := labels.SelectorFromSet(labelMap)

	return Eventually(
		func() int {
			pvcList, err := f.kubeClient.CoreV1().PersistentVolumeClaims(meta.Namespace).List(
				context.TODO(),
				metav1.ListOptions{
					LabelSelector: labelSelector.String(),
				},
			)
			Expect(err).NotTo(HaveOccurred())

			return len(pvcList.Items)
		},
		time.Minute*5,
		time.Second*5,
	)
}
