package e2e_test

import (
	"fmt"
	"strconv"
	"time"

	"github.com/appscode/go/types"
	"github.com/appscode/kutil/tools/portforward"
	rd "github.com/go-redis/redis"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/redis/test/e2e/framework"
	"github.com/kubedb/redis/test/e2e/matcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var createAndWaitForRunning = func() {
	By("Create RedisVersion: " + cl.redisVersion.Name)
	err := cl.f.CreateRedisVersion(cl.redisVersion)
	Expect(err).NotTo(HaveOccurred())

	By("Create Redis: " + cl.redis.Name)
	err = cl.f.CreateRedis(cl.redis)
	Expect(err).NotTo(HaveOccurred())

	By("Wait for Running redis")
	cl.f.EventuallyRedisRunning(cl.redis.ObjectMeta).Should(BeTrue())
}

var deleteTestResource = func() {
	By("Delete redis")
	err := cl.f.DeleteRedis(cl.redis.ObjectMeta)
	Expect(err).NotTo(HaveOccurred())

	By("Wait for redis to be paused")
	cl.f.EventuallyDormantDatabaseStatus(cl.redis.ObjectMeta).Should(matcher.HavePaused())

	By("WipeOut redis")
	_, err = cl.f.PatchDormantDatabase(cl.redis.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
		in.Spec.WipeOut = true
		return in
	})
	Expect(err).NotTo(HaveOccurred())

	By("Delete Dormant Database")
	err = cl.f.DeleteDormantDatabase(cl.redis.ObjectMeta)
	Expect(err).NotTo(HaveOccurred())

	By("Wait for redis resources to be wipedOut")
	cl.f.EventuallyWipedOut(cl.redis.ObjectMeta).Should(Succeed())

	By("Delete RedisVersion")
	err = cl.f.DeleteRedisVersion(cl.redisVersion.ObjectMeta)
	Expect(err).NotTo(HaveOccurred())
}

var _ = Describe("Redis Cluster", func() {
	var (
		err                  error
		skipMessage          string
		failover             bool
		opt                  *rd.ClusterOptions
		client               *rd.ClusterClient
		cluster              *framework.ClusterScenario
		ports                [][]string
		tunnels              [][]*portforward.Tunnel
		nodes                [][]framework.RedisNode
		rdClients            [][]*rd.Client
		expectedClusterSlots []rd.ClusterSlot
	)

	var clusterSlots = func() ([]rd.ClusterSlot, error) {
		var slots []rd.ClusterSlot

		for i := range nodes {
			for k := range nodes[i][0].SlotStart {
				slot := rd.ClusterSlot{
					Start: nodes[i][0].SlotStart[k],
					End:   nodes[i][0].SlotEnd[k],
					Nodes: make([]rd.ClusterNode, len(nodes[i])),
				}
				for j := 0; j < len(nodes[i]); j++ {
					slot.Nodes[j] = rd.ClusterNode{
						Addr: ":" + nodes[i][j].Port,
					}
				}

				slots = append(slots, slot)
			}
		}

		return slots, nil
	}

	var getConfiguredClusterInfo = func() {
		skipMessage = ""
		if !framework.Cluster {
			skipMessage = "cluster test is disabled"
		}

		By("Forward ports")
		ports, tunnels, err = cl.f.GetPodsIPWithTunnel(cl.redis)
		Expect(err).NotTo(HaveOccurred())

		By("Wait until redis cluster be configured")
		Expect(cl.f.WaitUntilRedisClusterConfigured(cl.redis, ports[0][0])).NotTo(HaveOccurred())

		By("Get configured cluster info")
		//cluster = framework.Sync(ports, cl.redis)
		nodes, rdClients = framework.Sync(ports, cl.redis)
		cluster = &framework.ClusterScenario{
			Nodes:   nodes,
			Clients: rdClients,
		}
	}

	var closeExistingTunnels = func() {
		By("closing tunnels")
		for i := range tunnels {
			for j := range tunnels[i] {
				tunnels[i][j].Close()
			}
		}
	}

	var createAndInitializeClusterClient = func() {
		By(fmt.Sprintf("Creating cluster client using ports %v", ports))
		opt = &rd.ClusterOptions{
			ClusterSlots:  clusterSlots,
			RouteRandomly: true,
		}
		client = cluster.ClusterClient(opt)
		Expect(client.ReloadState()).NotTo(HaveOccurred())

		By("Initializing cluster client")
		err := client.ForEachMaster(func(master *rd.Client) error {
			return master.FlushDB().Err()
		})
		Expect(err).NotTo(HaveOccurred())
	}

	var assertSimple = func() {
		It("should GET/SET/DEL", func() {
			if skipMessage != "" {
				Skip(skipMessage)
			}

			res := client.Get("A").Val()
			if failover {
				Expect(res).To(Equal("VALUE"))
			} else {
				Expect(res).To(Equal(""))
				err = client.Set("A", "VALUE", 0).Err()
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					return client.Get("A").Val()
				}, 30*time.Second).Should(Equal("VALUE"))
			}
		})
	}

	var assertPubSub = func() {
		It("supports PubSub", func() {
			if skipMessage != "" {
				Skip(skipMessage)
			}

			pubsub := client.Subscribe("mychannel")
			defer pubsub.Close()

			Eventually(func() error {
				_, err := client.Publish("mychannel", "hello").Result()
				if err != nil {
					return err
				}

				msg, err := pubsub.ReceiveTimeout(time.Second)
				if err != nil {
					return err
				}

				_, ok := msg.(*rd.Message)
				if !ok {
					return fmt.Errorf("got %T, wanted *redis.Message", msg)
				}

				return nil
			}, 30*time.Second).ShouldNot(HaveOccurred())
		})
	}

	Context("Cluster Commands", func() {
		BeforeEach(func() {
			getConfiguredClusterInfo()
			createAndInitializeClusterClient()
		})

		AfterEach(func() {
			err = client.ForEachMaster(func(master *rd.Client) error {
				return master.FlushDB().Err()
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(client.Close()).NotTo(HaveOccurred())

			closeExistingTunnels()
		})

		It("should CLUSTER INFO", func() {
			if skipMessage != "" {
				Skip(skipMessage)
			}

			res, err := client.ClusterInfo().Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(ContainSubstring(fmt.Sprintf("cluster_known_nodes:%d",
				(*cl.redis.Spec.Cluster.Master)*((*cl.redis.Spec.Cluster.Replicas)+1))))
		})

		It("calls fn for every master node", func() {
			if skipMessage != "" {
				Skip(skipMessage)
			}

			for i := 0; i < 10; i++ {
				Expect(client.Set(strconv.Itoa(i), "", 0).Err()).NotTo(HaveOccurred())
			}

			err := client.ForEachMaster(func(master *rd.Client) error {
				return master.FlushDB().Err()
			})
			Expect(err).NotTo(HaveOccurred())

			size, err := client.DBSize().Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(size).To(Equal(int64(0)))
		})

		It("should CLUSTER SLOTS", func() {
			if skipMessage != "" {
				Skip(skipMessage)
			}

			res, err := client.ClusterSlots().Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(HaveLen(3))

			wanted := []rd.ClusterSlot{
				{
					Start: 0,
					End:   5460,
					Nodes: cluster.ClusterNodes(0, 5460),
				}, {
					Start: 5461,
					End:   10922,
					Nodes: cluster.ClusterNodes(5461, 10922),
				}, {
					Start: 10923,
					End:   16383,
					Nodes: cluster.ClusterNodes(10923, 16383),
				},
			}

			Expect(framework.AssertSlotsEqual(res, wanted)).NotTo(HaveOccurred())
		})

		It("should CLUSTER NODES", func() {
			if skipMessage != "" {
				Skip(skipMessage)
			}

			res, err := client.ClusterNodes().Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(res)).To(BeNumerically(">", 400))
		})

		It("should CLUSTER COUNT-FAILURE-REPORTS", func() {
			if skipMessage != "" {
				Skip(skipMessage)
			}

			n, err := client.ClusterCountFailureReports(cluster.Nodes[0][0].ID).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(int64(0)))
		})

		It("should CLUSTER COUNTKEYSINSLOT", func() {
			if skipMessage != "" {
				Skip(skipMessage)
			}

			n, err := client.ClusterCountKeysInSlot(10).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(int64(0)))
		})

		It("should CLUSTER SAVECONFIG", func() {
			if skipMessage != "" {
				Skip(skipMessage)
			}

			res, err := client.ClusterSaveConfig().Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal("OK"))
		})

		It("should CLUSTER SLAVES", func() {
			if skipMessage != "" {
				Skip(skipMessage)
			}

			for i := range nodes {
				if nodes[i][0].Role == "master" {
					nodesList, err := client.ClusterSlaves(cluster.Nodes[i][0].ID).Result()
					Expect(err).NotTo(HaveOccurred())
					Expect(nodesList).Should(ContainElement(ContainSubstring("slave")))
					Expect(nodesList).Should(HaveLen(1))
					break
				}
			}
		})

		It("should RANDOMKEY", func() {
			if skipMessage != "" {
				Skip(skipMessage)
			}

			const nkeys = 100

			for i := 0; i < nkeys; i++ {
				err := client.Set(fmt.Sprintf("key%d", i), "value", 0).Err()
				Expect(err).NotTo(HaveOccurred())
			}

			var keys []string
			addKey := func(key string) {
				for _, k := range keys {
					if k == key {
						return
					}
				}
				keys = append(keys, key)
			}

			for i := 0; i < nkeys*10; i++ {
				key := client.RandomKey().Val()
				addKey(key)
			}

			Expect(len(keys)).To(BeNumerically("~", nkeys, nkeys/10))
		})

		assertSimple()
		assertPubSub()
	})

	Context("Cluster failover", func() {
		JustBeforeEach(func() {
			failover = true

			getConfiguredClusterInfo()
			createAndInitializeClusterClient()

			err = client.ForEachSlave(func(slave *rd.Client) error {
				defer GinkgoRecover()

				Eventually(func() int64 {
					return slave.DBSize().Val()
				}, "30s").Should(Equal(int64(0)))

				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			err = client.Set("A", "VALUE", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			err = client.ReloadState()
			Eventually(func() bool {
				err = client.ReloadState()
				if err != nil {
					return false
				}
				return true
			}, "30s").Should(BeTrue())

			err = client.ForEachSlave(func(slave *rd.Client) error {
				err = slave.ClusterFailover().Err()
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() bool {
					err := client.ReloadState()
					if err != nil {
						return false
					}
					return true
				}, "30s").Should(BeTrue())
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			failover = false

			Expect(client.Close()).NotTo(HaveOccurred())

			closeExistingTunnels()
		})

		assertSimple()
	})

	Context("Modify cluster", func() {
		It("should configure according to modified redis crd", func() {
			if skipMessage != "" {
				Skip(skipMessage)
			}

			By("Add replica")
			cl.redis, err = cl.f.TryPatchRedis(cl.redis.ObjectMeta, func(in *api.Redis) *api.Redis {
				in.Spec.Cluster.Replicas = types.Int32P((*cl.redis.Spec.Cluster.Replicas) + 1)
				return in
			})
			Expect(err).NotTo(HaveOccurred())

			By("Wait until statefulsets are ready")
			Expect(cl.f.WaitUntilStatefulSetReady(cl.redis)).NotTo(HaveOccurred())

			getConfiguredClusterInfo()
			createAndInitializeClusterClient()

			By("cluster slots should be configured as expected")
			expectedClusterSlots = []rd.ClusterSlot{
				{
					Start: 0,
					End:   5460,
					Nodes: cluster.ClusterNodes(0, 5460),
				}, {
					Start: 5461,
					End:   10922,
					Nodes: cluster.ClusterNodes(5461, 10922),
				}, {
					Start: 10923,
					End:   16383,
					Nodes: cluster.ClusterNodes(10923, 16383),
				},
			}
			Eventually(func() error {
				res, err := client.ClusterSlots().Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(HaveLen(3))

				return framework.AssertSlotsEqual(res, expectedClusterSlots)
			}, time.Minute*5, time.Second).ShouldNot(HaveOccurred())

			closeExistingTunnels()

			// =======================================

			By("Remove replica")
			cl.redis, err = cl.f.TryPatchRedis(cl.redis.ObjectMeta, func(in *api.Redis) *api.Redis {
				in.Spec.Cluster.Replicas = types.Int32P((*cl.redis.Spec.Cluster.Replicas) - 1)
				return in
			})
			Expect(err).NotTo(HaveOccurred())

			By("Wait until statefulsets are ready")
			Expect(cl.f.WaitUntilStatefulSetReady(cl.redis)).NotTo(HaveOccurred())

			getConfiguredClusterInfo()
			createAndInitializeClusterClient()

			By("cluster slots should be configured as expected")
			expectedClusterSlots = []rd.ClusterSlot{
				{
					Start: 0,
					End:   5460,
					Nodes: cluster.ClusterNodes(0, 5460),
				}, {
					Start: 5461,
					End:   10922,
					Nodes: cluster.ClusterNodes(5461, 10922),
				}, {
					Start: 10923,
					End:   16383,
					Nodes: cluster.ClusterNodes(10923, 16383),
				},
			}
			Eventually(func() error {
				res, err := client.ClusterSlots().Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(HaveLen(3))

				return framework.AssertSlotsEqual(res, expectedClusterSlots)
			}, time.Minute*5, time.Second).ShouldNot(HaveOccurred())

			closeExistingTunnels()

			// =======================================

			By("Add master")
			cl.redis, err = cl.f.TryPatchRedis(cl.redis.ObjectMeta, func(in *api.Redis) *api.Redis {
				in.Spec.Cluster.Master = types.Int32P((*cl.redis.Spec.Cluster.Master) + 1)
				return in
			})
			Expect(err).NotTo(HaveOccurred())

			By("Wait until statefulsets are ready")
			Expect(cl.f.WaitUntilStatefulSetReady(cl.redis)).NotTo(HaveOccurred())

			getConfiguredClusterInfo()
			createAndInitializeClusterClient()

			By("cluster slots should be configured as expected")
			expectedClusterSlots = []rd.ClusterSlot{
				{
					Start: 1365,
					End:   5460,
					Nodes: cluster.ClusterNodes(1365, 5460),
				}, {
					Start: 6827,
					End:   10922,
					Nodes: cluster.ClusterNodes(6827, 10922),
				}, {
					Start: 12288,
					End:   16383,
					Nodes: cluster.ClusterNodes(12288, 16383),
				}, {
					Start: 0,
					End:   1364,
					Nodes: cluster.ClusterNodes(0, 1364),
				}, {
					Start: 5461,
					End:   6826,
					Nodes: cluster.ClusterNodes(5461, 6826),
				}, {
					Start: 10923,
					End:   12287,
					Nodes: cluster.ClusterNodes(10923, 12287),
				},
			}
			Eventually(func() error {
				res, err := client.ClusterSlots().Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(HaveLen(6))

				return framework.AssertSlotsEqual(res, expectedClusterSlots)
			}, time.Minute*10, time.Second).ShouldNot(HaveOccurred())

			closeExistingTunnels()

			// =======================================

			By("Remove master")
			cl.redis, err = cl.f.TryPatchRedis(cl.redis.ObjectMeta, func(in *api.Redis) *api.Redis {
				in.Spec.Cluster.Master = types.Int32P((*cl.redis.Spec.Cluster.Master) - 1)
				return in
			})
			Expect(err).NotTo(HaveOccurred())

			By("Wait until statefulsets are ready")
			Expect(cl.f.WaitUntilStatefulSetReady(cl.redis)).NotTo(HaveOccurred())

			getConfiguredClusterInfo()
			createAndInitializeClusterClient()

			By("cluster slots should be configured as expected")
			expectedClusterSlots = []rd.ClusterSlot{
				{
					Start: 0,
					End:   5460,
					Nodes: cluster.ClusterNodes(0, 5460),
				}, {
					Start: 5461,
					End:   6825,
					Nodes: cluster.ClusterNodes(5461, 6825),
				}, {
					Start: 6827,
					End:   10922,
					Nodes: cluster.ClusterNodes(6827, 10922),
				}, {
					Start: 6826,
					End:   6826,
					Nodes: cluster.ClusterNodes(6826, 6826),
				}, {
					Start: 10923,
					End:   16383,
					Nodes: cluster.ClusterNodes(10923, 16383),
				},
			}
			Eventually(func() error {
				res, err := client.ClusterSlots().Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(HaveLen(5))

				return framework.AssertSlotsEqual(res, expectedClusterSlots)
			}, time.Minute*10, time.Second).ShouldNot(HaveOccurred())

			closeExistingTunnels()
		})
	})
})
