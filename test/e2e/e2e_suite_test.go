package e2e_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	cs "kubedb.dev/apimachinery/client/clientset/versioned"
	"kubedb.dev/apimachinery/client/clientset/versioned/scheme"
	"kubedb.dev/redis/test/e2e/framework"

	"github.com/appscode/go/log"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/kubernetes"
	clientSetScheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/util/homedir"
	ka "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
	"kmodules.xyz/client-go/logs"
	"kmodules.xyz/client-go/tools/clientcmd"
	appcat_cs "kmodules.xyz/custom-resources/client/clientset/versioned/typed/appcatalog/v1alpha1"
)

// To Run E2E tests:
//
// 1. ./hack/make.py test e2e
//
// 2. ./hack/make.py test e2e --v=1  --docker-registry=kubedbci --db-catalog=5.0 --db-version=5.0 --selfhosted-operator=true
type clusterVar struct {
	f     *framework.Invocation
	redis *api.Redis
}

var (
	storageClass   = "standard"
	kubeconfigPath = func() string {
		kubecfg := os.Getenv("KUBECONFIG")
		if kubecfg != "" {
			return kubecfg
		}
		return filepath.Join(homedir.HomeDir(), ".kube", "config")
	}()
	kubeContext = ""
)

func init() {
	utilruntime.Must(scheme.AddToScheme(clientSetScheme.Scheme))

	flag.StringVar(&kubeconfigPath, "kubeconfig", kubeconfigPath, "Path to kubeconfig file with authorization information (the master location is set by the master flag).")
	flag.StringVar(&kubeContext, "kube-context", "", "Name of kube context")
	flag.StringVar(&storageClass, "storageclass", storageClass, "Kubernetes StorageClass name")
	flag.StringVar(&framework.DockerRegistry, "docker-registry", framework.DockerRegistry, "User provided docker repository")
	flag.StringVar(&framework.DBCatalogName, "db-catalog", framework.DBCatalogName, "Postgres version")
	flag.BoolVar(&framework.SelfHostedOperator, "selfhosted-operator", framework.SelfHostedOperator, "Enable this for provided controller")
	flag.BoolVar(&framework.Cluster, "cluster", framework.Cluster, "Enable cluster tests")
}

const (
	TIMEOUT = 20 * time.Minute
)

var (
	root *framework.Framework
	cl   clusterVar
)

func TestE2e(t *testing.T) {
	logs.InitLogs()
	defer logs.FlushLogs()
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(TIMEOUT)

	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "e2e Suite", []Reporter{junitReporter})
}

var _ = BeforeSuite(func() {
	By("Using kubeconfig from " + kubeconfigPath)
	config, err := clientcmd.BuildConfigFromContext(kubeconfigPath, kubeContext)
	Expect(err).NotTo(HaveOccurred())
	// raise throttling time. ref: https://github.com/appscode/voyager/issues/640
	config.Burst = 100
	config.QPS = 100

	// Clients
	kubeClient := kubernetes.NewForConfigOrDie(config)
	extClient := cs.NewForConfigOrDie(config)
	kaClient := ka.NewForConfigOrDie(config)
	appCatalogClient, err := appcat_cs.NewForConfig(config)
	if err != nil {
		log.Fatalln(err)
	}
	// Framework
	root = framework.New(config, kubeClient, extClient, kaClient, appCatalogClient, storageClass)

	// Create namespace
	By("Using namespace " + root.Namespace())
	err = root.CreateNamespace()
	Expect(err).NotTo(HaveOccurred())

	if !framework.SelfHostedOperator {
		stopCh := genericapiserver.SetupSignalHandler()
		go root.RunOperatorAndServer(config, kubeconfigPath, stopCh)
	}

	root.EventuallyCRD().Should(Succeed())
	root.EventuallyAPIServiceReady().Should(Succeed())

	if framework.Cluster {
		cl = clusterVar{}
		cl.f = root.Invoke()
		cl.redis = cl.f.RedisCluster()
		createAndWaitForRunning()
	}
})

var _ = AfterSuite(func() {
	if framework.Cluster {
		deleteTestResource()
	}

	By("Cleanup Left Overs")
	if !framework.SelfHostedOperator {
		By("Delete Admission Controller Configs")
		root.CleanAdmissionConfigs()
	}
	By("Delete left over Redis objects")
	root.CleanRedis()
	By("Delete left over Dormant Database objects")
	root.CleanDormantDatabase()
	By("Delete Namespace")
	err := root.DeleteNamespace()
	Expect(err).NotTo(HaveOccurred())
})
