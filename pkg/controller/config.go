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

	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	amc "kubedb.dev/apimachinery/pkg/controller"
	"kubedb.dev/apimachinery/pkg/eventer"

	pcm "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	reg_util "kmodules.xyz/client-go/admissionregistration/v1beta1"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/discovery"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned"
)

const (
	mutatingWebhookConfig   = "mutators.kubedb.com"
	validatingWebhookConfig = "validators.kubedb.com"
)

type OperatorConfig struct {
	amc.Config

	ClientConfig     *rest.Config
	KubeClient       kubernetes.Interface
	APIExtKubeClient crd_cs.ApiextensionsV1beta1Interface
	DBClient         cs.Interface
	DynamicClient    dynamic.Interface
	AppCatalogClient appcat_cs.Interface
	PromClient       pcm.MonitoringV1Interface
}

func NewOperatorConfig(clientConfig *rest.Config) *OperatorConfig {
	return &OperatorConfig{
		ClientConfig: clientConfig,
	}
}

func (c *OperatorConfig) New() (*Controller, error) {
	if err := discovery.IsDefaultSupportedVersion(c.KubeClient); err != nil {
		return nil, err
	}

	topology, err := core_util.DetectTopology(context.TODO(), c.KubeClient)
	if err != nil {
		return nil, err
	}

	recorder := eventer.NewEventRecorder(c.KubeClient, "Redis operator")

	ctrl := New(
		c.ClientConfig,
		c.KubeClient,
		c.APIExtKubeClient,
		c.DBClient,
		c.DynamicClient,
		c.AppCatalogClient,
		c.PromClient,
		c.Config,
		topology,
		recorder,
	)

	if err := ctrl.EnsureCustomResourceDefinitions(); err != nil {
		return nil, err
	}
	if c.EnableMutatingWebhook {
		if err := reg_util.UpdateMutatingWebhookCABundle(c.ClientConfig, mutatingWebhookConfig); err != nil {
			return nil, err
		}
	}
	if c.EnableValidatingWebhook {
		if err := reg_util.UpdateValidatingWebhookCABundle(c.ClientConfig, validatingWebhookConfig); err != nil {
			return nil, err
		}
	}

	if err := ctrl.Init(); err != nil {
		return nil, err
	}

	return ctrl, nil
}
