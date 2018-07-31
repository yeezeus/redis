package e2e_test

import (
	"fmt"

	exec_util "github.com/appscode/kutil/tools/exec"
	api "github.com/kubedb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/kubedb/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
	"github.com/kubedb/redis/test/e2e/framework"
	"github.com/kubedb/redis/test/e2e/matcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
)

var _ = Describe("Redis", func() {
	var (
		err         error
		f           *framework.Invocation
		redis       *api.Redis
		skipMessage string
	)

	BeforeEach(func() {
		f = root.Invoke()
		redis = f.Redis()
		skipMessage = ""
	})

	var createAndWaitForRunning = func() {
		By("Create Redis: " + redis.Name)
		err = f.CreateRedis(redis)
		Expect(err).NotTo(HaveOccurred())

		By("Wait for Running redis")
		f.EventuallyRedisRunning(redis.ObjectMeta).Should(BeTrue())
	}

	var deleteTestResource = func() {
		By("Delete redis")
		err = f.DeleteRedis(redis.ObjectMeta)
		Expect(err).NotTo(HaveOccurred())

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

		By("Wait for redis resources to be wipedOut")
		f.EventuallyWipedOut(redis.ObjectMeta).Should(Succeed())
	}

	var shouldSuccessfullyRunning = func() {
		if skipMessage != "" {
			Skip(skipMessage)
		}

		// Create Redis
		createAndWaitForRunning()

		// Delete test resource
		deleteTestResource()
	}

	Describe("Test", func() {

		Context("General", func() {

			Context("-", func() {
				It("should run successfully", shouldSuccessfullyRunning)
			})
		})

		Context("DoNotPause", func() {
			BeforeEach(func() {
				redis.Spec.DoNotPause = true
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

				By("Update redis to set DoNotPause=false")
				f.TryPatchRedis(redis.ObjectMeta, func(in *api.Redis) *api.Redis {
					in.Spec.DoNotPause = false
					return in
				})

				// Delete test resource
				deleteTestResource()
			})
		})

		Context("Resume", func() {
			var usedInitSpec bool
			BeforeEach(func() {
				usedInitSpec = false
			})

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

					// Delete test resource
					deleteTestResource()
				})
			})

			Context("Basic Resume", func() {
				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running Redis
					createAndWaitForRunning()
					By("Delete redis")
					f.DeleteRedis(redis.ObjectMeta)

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

					_, err = f.GetRedis(redis.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					// Delete test resource
					deleteTestResource()
				})
			})

			Context("Multiple times with PVC", func() {
				It("should resume DormantDatabase successfully", func() {
					// Create and wait for running Redis
					createAndWaitForRunning()

					for i := 0; i < 3; i++ {
						By(fmt.Sprintf("%v-th", i+1) + " time running.")
						By("Delete redis")
						f.DeleteRedis(redis.ObjectMeta)

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

						_, err := f.GetRedis(redis.ObjectMeta)
						Expect(err).NotTo(HaveOccurred())
					}

					// Delete test resource
					deleteTestResource()
				})
			})
		})

		Context("Environment Variables", func() {
			AfterEach(func() {
				deleteTestResource()
			})

			envList := []core.EnvVar{
				{
					Name:  "TEST_ENV",
					Value: "kubedb-redis-e2e",
				},
			}

			Context("Allowed Envs", func() {
				It("should run successfully with given Env", func() {
					redis.Spec.Env = envList
					createAndWaitForRunning()

					By("Checking pod started with given envs")
					pod, err := f.GetPod(redis.ObjectMeta)
					Expect(err).NotTo(HaveOccurred())

					out, err := exec_util.ExecIntoPod(f.RestConfig(), pod, "env")
					Expect(err).NotTo(HaveOccurred())
					for _, env := range envList {
						Expect(out).Should(ContainSubstring(env.Name + "=" + env.Value))
					}

				})
			})

			Context("Update Envs", func() {
				It("should reject to update Env", func() {
					redis.Spec.Env = envList
					createAndWaitForRunning()

					By("Updating Envs")
					_, _, err := util.PatchRedis(f.ExtClient(), redis, func(in *api.Redis) *api.Redis {
						in.Spec.Env = []core.EnvVar{
							{
								Name:  "TEST_ENV",
								Value: "patched",
							},
						}
						return in
					})

					Expect(err).To(HaveOccurred())
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
					testSvc    *core.Service
				)

				BeforeEach(func() {
					userConfig = f.GetCustomConfig(customConfigs)
					testSvc = f.GetTestService(redis.ObjectMeta)

					By("Creating Service: " + testSvc.Name)
					f.CreateService(testSvc)
				})

				AfterEach(func() {
					By("Deleting configMap: " + userConfig.Name)
					f.DeleteConfigMap(userConfig.ObjectMeta)

					By("Deleting Service: " + testSvc.Name)
					f.DeleteService(testSvc.ObjectMeta)
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
	})
})
