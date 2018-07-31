package framework

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) GetDatabasePod(meta metav1.ObjectMeta) (*core.Pod, error) {
	return f.kubeClient.CoreV1().Pods(meta.Namespace).Get(meta.Name+"-0", metav1.GetOptions{})
}

func (f *Framework) GetRedisClient(meta metav1.ObjectMeta) (*redis.Client, error) {
	clusterIP := net.IP{192, 168, 99, 100} //minikube ip

	pod, err := f.GetDatabasePod(meta)
	if err != nil {
		return nil, err
	}

	if pod.Spec.NodeName != "minikube" {
		node, err := f.kubeClient.CoreV1().Nodes().Get(pod.Spec.NodeName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}

		for _, addr := range node.Status.Addresses {
			if addr.Type == core.NodeExternalIP {
				clusterIP = net.ParseIP(addr.Address)
				break
			}
		}
	}

	svc, err := f.kubeClient.CoreV1().Services(f.Namespace()).Get(meta.Name+"-test-svc", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	nodePort := strconv.Itoa(int(svc.Spec.Ports[0].NodePort))
	address := fmt.Sprintf(clusterIP.String() + ":" + nodePort)

	return redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "", // no password set
		DB:       0,  // use default DB
	}), nil
}

func (f *Framework) EventuallyRedisConfig(meta metav1.ObjectMeta, config string) GomegaAsyncAssertion {
	configPair := strings.Split(config, " ")

	return Eventually(
		func() string {

			client, err := f.GetRedisClient(meta)
			Expect(err).NotTo(HaveOccurred())

			// ping database to check if it is ready
			pong, err := client.Ping().Result()
			if err != nil {
				return ""
			}

			if !strings.Contains(pong, "PONG") {
				return ""
			}

			// get configuration
			response := client.ConfigGet(configPair[0])
			result := response.Val()
			ret := make([]string, 0)
			for _, r := range result {
				ret = append(ret, r.(string))
			}
			return strings.Join(ret, " ")
		},
		time.Minute*5,
		time.Second*5,
	)
}
