package framework

import (
	"fmt"

	api "github.com/kubedb/apimachinery/apis/catalog/v1alpha1"
	. "github.com/onsi/gomega"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (i *Invocation) RedisVersion() *api.RedisVersion {
	return &api.RedisVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: DBCatalogName,
			Labels: map[string]string{
				"app": i.app,
			},
		},
		Spec: api.RedisVersionSpec{
			Version: DBVersion,
			DB: api.RedisVersionDatabase{
				Image: fmt.Sprintf("%s/redis:%s", DockerRegistry, DBImageTag),
			},
			Exporter: api.RedisVersionExporter{
				Image: fmt.Sprintf("%s/redis_exporter:%s", DockerRegistry, ExporterTag),
			},
			PodSecurityPolicies: api.RedisVersionPodSecurityPolicy{
				DatabasePolicyName: "redis-db",
			},
		},
	}
}

func (f *Framework) CreateRedisVersion(obj *api.RedisVersion) error {
	_, err := f.extClient.CatalogV1alpha1().RedisVersions().Create(obj)
	if err != nil && kerr.IsAlreadyExists(err) {
		e2 := f.extClient.CatalogV1alpha1().RedisVersions().Delete(obj.Name, &metav1.DeleteOptions{})
		Expect(e2).NotTo(HaveOccurred())
		_, e2 = f.extClient.CatalogV1alpha1().RedisVersions().Create(obj)
		return e2
	}
	return nil
}

func (f *Framework) DeleteRedisVersion(meta metav1.ObjectMeta) error {
	return f.extClient.CatalogV1alpha1().RedisVersions().Delete(meta.Name, &metav1.DeleteOptions{})
}
