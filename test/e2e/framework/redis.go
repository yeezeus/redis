package framework

import (
	"time"

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/encoding/json/types"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/typed/kubedb/v1alpha1/util"
	. "github.com/onsi/gomega"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Invocation) Redis() *api.Redis {
	return &api.Redis{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rand.WithUniqSuffix("redis"),
			Namespace: f.namespace,
			Labels: map[string]string{
				"app": f.app,
			},
		},
		Spec: api.RedisSpec{
			Version: types.StrYo("4-alpine"),
		},
	}
}

func (f *Framework) CreateRedis(obj *api.Redis) error {
	_, err := f.extClient.Redises(obj.Namespace).Create(obj)
	return err
}

func (f *Framework) GetRedis(meta metav1.ObjectMeta) (*api.Redis, error) {
	return f.extClient.Redises(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
}

func (f *Framework) TryPatchRedis(meta metav1.ObjectMeta, transform func(*api.Redis) *api.Redis) (*api.Redis, error) {
	return util.TryPatchRedis(f.extClient, meta, transform)
}

func (f *Framework) DeleteRedis(meta metav1.ObjectMeta) error {
	return f.extClient.Redises(meta.Namespace).Delete(meta.Name, &metav1.DeleteOptions{})
}

func (f *Framework) EventuallyRedis(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			_, err := f.extClient.Redises(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			if err != nil {
				if kerr.IsNotFound(err) {
					return false
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			}
			return true
		},
		time.Minute*5,
		time.Second*5,
	)
}

func (f *Framework) EventuallyRedisRunning(meta metav1.ObjectMeta) GomegaAsyncAssertion {
	return Eventually(
		func() bool {
			redis, err := f.extClient.Redises(meta.Namespace).Get(meta.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
			return redis.Status.Phase == api.DatabasePhaseRunning
		},
		time.Minute*5,
		time.Second*5,
	)
}
