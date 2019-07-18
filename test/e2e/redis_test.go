package e2e_test

import (
	"fmt"

	"github.com/appscode/go/crypto/rand"
	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	exec_util "kmodules.xyz/client-go/tools/exec"
	catalog "kubedb.dev/apimachinery/apis/catalog/v1alpha1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"kubedb.dev/redis/test/e2e/framework"
	"kubedb.dev/redis/test/e2e/matcher"
)

var _ = Describe("Redis", func() {
	var (
		err          error
		f            *framework.Invocation
		redis        *api.Redis
		redisVersion *catalog.RedisVersion
		skipMessage  string
		key          string
		value        string
	)

	BeforeEach(func() {
		f = root.Invoke()
		redis = f.Redis()
		redisVersion = f.RedisVersion()
		skipMessage = ""
		key = rand.WithUniqSuffix("kubed-e2e")
		value = rand.GenerateTokenWithLength(10)
	})

	var createAndWaitForRunning = func() {
		By("Create RedisVersion: " + redisVersion.Name)
		err = f.CreateRedisVersion(redisVersion)
		Expect(err).NotTo(HaveOccurred())

		By("Create Redis: " + redis.Name)
		err = f.CreateRedis(redis)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for Running redis")
		f.EventuallyRedisRunning(redis.ObjectMeta).Should(BeTrue())

		By("Wait for AppBinding to create")
		f.EventuallyAppBinding(redis.ObjectMeta).Should(BeTrue())

		By("Check valid AppBinding Specs")
		err := f.CheckAppBindingSpec(redis.ObjectMeta)
		Expect(err).NotTo(HaveOccurred())
	}

	var deleteTestResource = func() {
		if redis == nil {
			// No redis. So, no cleanup
			return
		}

		By("Check if Redis " + redis.Name + " exists.")
		rd, err := f.GetRedis(redis.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				// Redis was not created. Hence, rest of cleanup is not necessary.
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		By("Delete redis")
		err = f.DeleteRedis(redis.ObjectMeta)
		if err != nil {
			if kerr.IsNotFound(err) {
				// Redis was not created. Hence, rest of cleanup is not necessary.
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}

		if rd.Spec.TerminationPolicy == api.TerminationPolicyPause {

			By("Wait for redis to be paused")
			f.EventuallyDormantDatabaseStatus(redis.ObjectMeta).Should(matcher.HavePaused())

			By("WipeOut redis")
			_, err := f.PatchDormantDatabase(redis.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
				in.Spec.WipeOut = true
				return in
			})
			Expect(err).NotTo(HaveOccurred())

			By("Delete Dormant Database")
			err = f.DeleteDormantDatabase(redis.ObjectMeta)
			Expect(err).NotTo(HaveOccurred())
		}

		By("Wait for redis resources to be wipedOut")
		f.EventuallyWipedOut(redis.ObjectMeta).Should(Succeed())
	}

	AfterEach(func() {

		deleteTestResource()

		By("Delete RedisVersion")
		err = f.DeleteRedisVersion(redisVersion.ObjectMeta)
		if err != nil && !kerr.IsNotFound(err) {
			Expect(err).NotTo(HaveOccurred())
		}
	})

	var shouldSuccessfullyRunning = func() {
		if skipMessage != "" {
			Skip(skipMessage)
		}

		// Create Redis
		createAndWaitForRunning()

		By("Inserting item into database")
		f.EventuallySetItem(redis.ObjectMeta, key, value).Should(BeTrue())

		By("Retrieving item from database")
		f.EventuallyGetItem(redis.ObjectMeta, key).Should(BeEquivalentTo(value))
	}

	Describe("Test", func() {

		Context("General", func() {

			Context("-", func() {
				It("should run successfully", func() {

					shouldSuccessfullyRunning()

					By("Delete redis")
					err = f.DeleteRedis(redis.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for redis to be paused")
					f.EventuallyDormantDatabaseStatus(redis.ObjectMeta).Should(matcher.HavePaused())

					// Create Redis object again to resume it
					By("Create Redis: " + redis.Name)
					err = f.CreateRedis(redis)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(redis.ObjectMeta).Should(BeFalse())

					By("Wait for Running redis")
					f.EventuallyRedisRunning(redis.ObjectMeta).Should(BeTrue())

					By("Retrieving item from database")
					f.EventuallyGetItem(redis.ObjectMeta, key).Should(BeEquivalentTo(value))

				})
			})

			Context("PDB", func() {

				It("should run eviction successfully", func() {
					// Create Redis
					By("Create DB")
					createAndWaitForRunning()
					//Evict Redis pod
					By("Try to evict Pod")
					err := f.EvictPodsFromStatefulSet(redis.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should run eviction with shard successfully", func() {
					redis = f.RedisCluster()
					redis.Spec.Cluster = &api.RedisClusterSpec{
						Master:   types.Int32P(3),
						Replicas: types.Int32P(2),
					}
					// Create Redis
					By("Create DB")
					createAndWaitForRunning()
					//Evict a Redis pod from each sts and deploy
					By("Try to evict Pod")
					err := f.EvictPodsFromStatefulSet(redis.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("Custom Resources", func() {

				Context("with custom SA Name", func() {
					BeforeEach(func() {
						redis.Spec.PodTemplate.Spec.ServiceAccountName = "my-custom-sa"
						redis.Spec.TerminationPolicy = api.TerminationPolicyPause
					})

					It("should start and resume successfully", func() {
						//shouldTakeSnapshot()
						createAndWaitForRunning()
						By("Check if Postgres " + redis.Name + " exists.")
						_, err := f.GetRedis(redis.ObjectMeta)
						if err != nil {
							if kerr.IsNotFound(err) {
								// Postgres was not created. Hence, rest of cleanup is not necessary.
								return
							}
							Expect(err).NotTo(HaveOccurred())
						}

						By("Delete redis: " + redis.Name)
						err = f.DeleteRedis(redis.ObjectMeta)
						if err != nil {
							if kerr.IsNotFound(err) {
								// Postgres was not created. Hence, rest of cleanup is not necessary.
								log.Infof("Skipping rest of cleanup. Reason: Postgres %s is not found.", redis.Name)
								return
							}
							Expect(err).NotTo(HaveOccurred())
						}

						By("Wait for redis to be paused")
						f.EventuallyDormantDatabaseStatus(redis.ObjectMeta).Should(matcher.HavePaused())

						By("Resume DB")
						createAndWaitForRunning()
					})
				})

				Context("with custom SA", func() {
					var customSAForDB *core.ServiceAccount
					var customRoleForDB *rbac.Role
					var customRoleBindingForDB *rbac.RoleBinding
					BeforeEach(func() {
						customSAForDB = f.ServiceAccount()
						redis.Spec.PodTemplate.Spec.ServiceAccountName = customSAForDB.Name
						customRoleForDB = f.RoleForElasticsearch(redis.ObjectMeta)
						customRoleBindingForDB = f.RoleBinding(customSAForDB.Name, customRoleForDB.Name)
					})
					It("should take snapshot successfully", func() {
						By("Create Database SA")
						err = f.CreateServiceAccount(customSAForDB)
						Expect(err).NotTo(HaveOccurred())

						By("Create Database Role")
						err = f.CreateRole(customRoleForDB)
						Expect(err).NotTo(HaveOccurred())

						By("Create Database RoleBinding")
						err = f.CreateRoleBinding(customRoleBindingForDB)
						Expect(err).NotTo(HaveOccurred())

						createAndWaitForRunning()
					})
				})
			})
		})

		Context("Resume", func() {

			Context("Super Fast User - Create-Delete-Create-Delete-Create ", func() {
				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running Redis
					createAndWaitForRunning()

					By("Delete redis")
					err = f.DeleteRedis(redis.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for redis to be paused")
					f.EventuallyDormantDatabaseStatus(redis.ObjectMeta).Should(matcher.HavePaused())

					// Create Redis object again to resume it
					By("Create Redis: " + redis.Name)
					err = f.CreateRedis(redis)
					Expect(err).NotTo(HaveOccurred())

					// Delete without caring if DB is resumed
					By("Delete redis")
					err = f.DeleteRedis(redis.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for Redis to be paused")
					f.EventuallyRedis(redis.ObjectMeta).Should(BeFalse())

					// Create Redis object again to resume it
					By("Create Redis: " + redis.Name)
					err = f.CreateRedis(redis)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(redis.ObjectMeta).Should(BeFalse())

					By("Wait for Running redis")
					f.EventuallyRedisRunning(redis.ObjectMeta).Should(BeTrue())

					_, err = f.GetRedis(redis.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("Basic Resume", func() {
				It("should resume DormantDatabase successfully", func() {

					shouldSuccessfullyRunning()

					By("Delete redis")
					err = f.DeleteRedis(redis.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for redis to be paused")
					f.EventuallyDormantDatabaseStatus(redis.ObjectMeta).Should(matcher.HavePaused())

					// Create Redis object again to resume it
					By("Create Redis: " + redis.Name)
					err = f.CreateRedis(redis)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(redis.ObjectMeta).Should(BeFalse())

					By("Wait for Running redis")
					f.EventuallyRedisRunning(redis.ObjectMeta).Should(BeTrue())

					By("Retrieving item from database")
					f.EventuallyGetItem(redis.ObjectMeta, key).Should(BeEquivalentTo(value))
				})
			})

			Context("Multiple times with PVC", func() {
				It("should resume DormantDatabase successfully", func() {

					shouldSuccessfullyRunning()

					for i := 0; i < 3; i++ {
						By(fmt.Sprintf("%v-th", i+1) + " time running.")
						By("Delete redis")
						err = f.DeleteRedis(redis.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())

						By("Wait for redis to be paused")
						f.EventuallyDormantDatabaseStatus(redis.ObjectMeta).Should(matcher.HavePaused())

						// Create Redis object again to resume it
						By("Create Redis: " + redis.Name)
						err = f.CreateRedis(redis)
						Expect(err).NotTo(HaveOccurred())

						By("Wait for DormantDatabase to be deleted")
						f.EventuallyDormantDatabase(redis.ObjectMeta).Should(BeFalse())

						By("Wait for Running redis")
						f.EventuallyRedisRunning(redis.ObjectMeta).Should(BeTrue())

						By("Retrieving item from database")
						f.EventuallyGetItem(redis.ObjectMeta, key).Should(BeEquivalentTo(value))
					}
				})
			})
		})

		Context("Termination Policy", func() {

			Context("with TerminationPolicyDoNotTerminate", func() {
				BeforeEach(func() {
					redis.Spec.TerminationPolicy = api.TerminationPolicyDoNotTerminate
				})

				It("should work successfully", func() {
					// Create and wait for running Redis
					createAndWaitForRunning()

					By("Delete redis")
					err = f.DeleteRedis(redis.ObjectMeta)
					Expect(err).Should(HaveOccurred())

					By("Redis is not paused. Check for redis")
					f.EventuallyRedis(redis.ObjectMeta).Should(BeTrue())

					By("Check for Running redis")
					f.EventuallyRedisRunning(redis.ObjectMeta).Should(BeTrue())

					By("Update redis to set spec.terminationPolicy = Pause")
					_, err := f.TryPatchRedis(redis.ObjectMeta, func(in *api.Redis) *api.Redis {
						in.Spec.TerminationPolicy = api.TerminationPolicyPause
						return in
					})
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("with TerminationPolicyPause (default)", func() {
				var shouldRunWithTerminationPause = func() {

					shouldSuccessfullyRunning()

					By("Delete redis")
					err = f.DeleteRedis(redis.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					// DormantDatabase.Status= paused, means redis object is deleted
					By("Wait for redis to be paused")
					f.EventuallyDormantDatabaseStatus(redis.ObjectMeta).Should(matcher.HavePaused())

					// Create Redis object again to resume it
					By("Create (pause) Redis: " + redis.Name)
					err = f.CreateRedis(redis)
					Expect(err).NotTo(HaveOccurred())

					By("Wait for DormantDatabase to be deleted")
					f.EventuallyDormantDatabase(redis.ObjectMeta).Should(BeFalse())

					By("Wait for Running redis")
					f.EventuallyRedisRunning(redis.ObjectMeta).Should(BeTrue())

					By("Retrieving item from database")
					f.EventuallyGetItem(redis.ObjectMeta, key).Should(BeEquivalentTo(value))

				}

				It("should create dormantdatabase successfully", shouldRunWithTerminationPause)
			})

			Context("with TerminationPolicyDelete", func() {
				BeforeEach(func() {
					redis.Spec.TerminationPolicy = api.TerminationPolicyDelete
				})

				var shouldRunWithTerminationDelete = func() {

					shouldSuccessfullyRunning()

					By("Delete redis")
					err = f.DeleteRedis(redis.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("wait until redis is deleted")
					f.EventuallyRedis(redis.ObjectMeta).Should(BeFalse())

					By("Checking DormantDatabase is not created")
					f.EventuallyDormantDatabase(redis.ObjectMeta).Should(BeFalse())

					By("Check for deleted PVCs")
					f.EventuallyPVCCount(redis.ObjectMeta).Should(Equal(0))
				}

				It("should run with TerminationPolicyDelete", shouldRunWithTerminationDelete)
			})

			Context("with TerminationPolicyWipeOut", func() {
				BeforeEach(func() {
					redis.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
				})

				var shouldRunWithTerminationWipeOut = func() {

					shouldSuccessfullyRunning()

					By("Delete redis")
					err = f.DeleteRedis(redis.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					By("wait until redis is deleted")
					f.EventuallyRedis(redis.ObjectMeta).Should(BeFalse())

					By("Checking DormantDatabase is not created")
					f.EventuallyDormantDatabase(redis.ObjectMeta).Should(BeFalse())

					By("Check for deleted PVCs")
					f.EventuallyPVCCount(redis.ObjectMeta).Should(Equal(0))
				}

				It("should run with TerminationPolicyDelete", shouldRunWithTerminationWipeOut)
			})
		})

		Context("Environment Variables", func() {

			envList := []core.EnvVar{
				{
					Name:  "TEST_ENV",
					Value: "kubedb-redis-e2e",
				},
			}

			Context("Allowed Envs", func() {
				It("should run successfully with given Env", func() {
					redis.Spec.PodTemplate.Spec.Env = envList
					createAndWaitForRunning()

					By("Checking pod started with given envs")
					pod, err := f.GetPod(redis.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					out, err := exec_util.ExecIntoPod(f.RestConfig(), pod, exec_util.Command("env"))
					Expect(err).NotTo(HaveOccurred())
					for _, env := range envList {
						Expect(out).Should(ContainSubstring(env.Name + "=" + env.Value))
					}

				})
			})

			Context("Update Envs", func() {
				It("should not reject to update Env", func() {
					redis.Spec.PodTemplate.Spec.Env = envList
					createAndWaitForRunning()

					By("Updating Envs")
					_, _, err := util.PatchRedis(f.ExtClient().KubedbV1alpha1(), redis, func(in *api.Redis) *api.Redis {
						in.Spec.PodTemplate.Spec.Env = []core.EnvVar{
							{
								Name:  "TEST_ENV",
								Value: "patched",
							},
						}
						return in
					})

					Expect(err).NotTo(HaveOccurred())
				})
			})

		})

		Context("Custom config", func() {

			customConfigs := []string{
				"databases 10",
				"maxclients 500",
			}

			Context("from configMap", func() {
				var (
					userConfig *core.ConfigMap
				)

				BeforeEach(func() {
					userConfig = f.GetCustomConfig(customConfigs)
				})

				AfterEach(func() {
					By("Deleting configMap: " + userConfig.Name)
					err := f.DeleteConfigMap(userConfig.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should set configuration provided in configMap", func() {
					if skipMessage != "" {
						Skip(skipMessage)
					}

					By("Creating configMap: " + userConfig.Name)
					err := f.CreateConfigMap(userConfig)
					Expect(err).NotTo(HaveOccurred())

					redis.Spec.ConfigSource = &core.VolumeSource{
						ConfigMap: &core.ConfigMapVolumeSource{
							LocalObjectReference: core.LocalObjectReference{
								Name: userConfig.Name,
							},
						},
					}

					// Create Redis
					createAndWaitForRunning()

					By("Checking redis configured from provided custom configuration")
					for _, cfg := range customConfigs {
						f.EventuallyRedisConfig(redis.ObjectMeta, cfg).Should(matcher.UseCustomConfig(cfg))
					}
				})
			})

		})

		Context("StorageType ", func() {

			var shouldRunSuccessfully = func() {

				if skipMessage != "" {
					Skip(skipMessage)
				}

				// Create Redis
				createAndWaitForRunning()

				By("Inserting item into database")
				f.EventuallySetItem(redis.ObjectMeta, key, value).Should(BeTrue())

				By("Retrieving item from database")
				f.EventuallyGetItem(redis.ObjectMeta, key).Should(BeEquivalentTo(value))
			}

			Context("Ephemeral", func() {

				BeforeEach(func() {
					redis.Spec.StorageType = api.StorageTypeEphemeral
					redis.Spec.Storage = nil
				})

				Context("General Behaviour", func() {

					BeforeEach(func() {
						redis.Spec.TerminationPolicy = api.TerminationPolicyWipeOut
					})

					It("should run successfully", shouldRunSuccessfully)
				})

				Context("With TerminationPolicyPause", func() {

					BeforeEach(func() {
						redis.Spec.TerminationPolicy = api.TerminationPolicyPause
					})

					It("should reject to create Redis object", func() {

						By("Creating Redis: " + redis.Name)
						err := f.CreateRedis(redis)
						Expect(err).To(HaveOccurred())
					})
				})
			})
		})
	})
})
