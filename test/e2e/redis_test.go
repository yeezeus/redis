package e2e_test

import (
	"fmt"

	"github.com/appscode/go/types"
	api "github.com/k8sdb/apimachinery/apis/kubedb/v1alpha1"
	"github.com/k8sdb/redis/test/e2e/framework"
	"github.com/k8sdb/redis/test/e2e/matcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
		_, err := f.TryPatchDormantDatabase(redis.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
			in.Spec.WipeOut = true
			return in
		})
		Expect(err).NotTo(HaveOccurred())

		By("Wait for redis to be wipedOut")
		f.EventuallyDormantDatabaseStatus(redis.ObjectMeta).Should(matcher.HaveWipedOut())

		err = f.DeleteDormantDatabase(redis.ObjectMeta)
		Expect(err).NotTo(HaveOccurred())
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

			Context("With PVC", func() {
				BeforeEach(func() {
					if f.StorageClass == "" {
						skipMessage = "Missing StorageClassName. Provide as flag to test this."
					}
					redis.Spec.Storage = &core.PersistentVolumeClaimSpec{
						Resources: core.ResourceRequirements{
							Requests: core.ResourceList{
								core.ResourceStorage: resource.MustParse("100Mi"),
							},
						},
						StorageClassName: types.StringP(f.StorageClass),
					}
				})
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
				Expect(err).NotTo(HaveOccurred())

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

			var shouldResumeSuccessfully = func() {
				// Create and wait for running Redis
				createAndWaitForRunning()

				By("Delete redis")
				f.DeleteRedis(redis.ObjectMeta)

				By("Wait for redis to be paused")
				f.EventuallyDormantDatabaseStatus(redis.ObjectMeta).Should(matcher.HavePaused())

				_, err = f.TryPatchDormantDatabase(redis.ObjectMeta, func(in *api.DormantDatabase) *api.DormantDatabase {
					in.Spec.Resume = true
					return in
				})
				Expect(err).NotTo(HaveOccurred())

				By("Wait for DormantDatabase to be deleted")
				f.EventuallyDormantDatabase(redis.ObjectMeta).Should(BeFalse())

				By("Wait for Running redis")
				f.EventuallyRedisRunning(redis.ObjectMeta).Should(BeTrue())

				redis, err = f.GetRedis(redis.ObjectMeta)
				Expect(err).NotTo(HaveOccurred())

				// Delete test resource
				deleteTestResource()
			}

			Context("-", func() {
				It("should resume DormantDatabase successfully", shouldResumeSuccessfully)
			})

			Context("With original Redis", func() {
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

				Context("Multiple times with PVC", func() {
					BeforeEach(func() {
						if f.StorageClass == "" {
							skipMessage = "Missing StorageClassName. Provide as flag to test this."
						}
						redis.Spec.Storage = &core.PersistentVolumeClaimSpec{
							Resources: core.ResourceRequirements{
								Requests: core.ResourceList{
									core.ResourceStorage: resource.MustParse("100Mi"),
								},
							},
							StorageClassName: types.StringP(f.StorageClass),
						}
					})

					By("Using a false init script.")

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
		})

	})
})
