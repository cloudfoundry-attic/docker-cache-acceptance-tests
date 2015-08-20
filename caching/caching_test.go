package caching

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/cloudfoundry-incubator/docker-cache-acceptance-tests/commons"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("A Docker App", func() {
	var appName string
	var registryAddress string

	BeforeEach(func() {
		appName = generator.RandomName()
		registryAddress = helpers.LoadConfig().DockerRegistryAddress
	})

	Context("pushed to Diego", func() {
		JustBeforeEach(func() {
			Eventually(cf.Cf("docker-push", appName, "cloudfoundry/diego-docker-app:latest", "--no-start")).Should(Exit(0))
		})

		AfterEach(func() {
			Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
			Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
		})

		Context("with caching enabled", func() {

			JustBeforeEach(func() {
				Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "true")).Should(Exit(0))
				Eventually(cf.Cf("start", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
				Eventually(helpers.CurlingAppRoot(appName)).Should(Equal(OK_RESPONSE))
			})

			It("has its public image cached in the private registry", func() {
				AssertImageAvailable(registryAddress, getAppImageName(appName))
			})
		})

		Context("with caching disabled", func() {

			JustBeforeEach(func() {
				Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "false")).Should(Exit(0))
				Eventually(cf.Cf("start", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
				Eventually(helpers.CurlingAppRoot(appName)).Should(Equal(OK_RESPONSE))
			})

			Context("and then restaged with caching enabled", func() {

				JustBeforeEach(func() {
					Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "true")).Should(Exit(0))
					Eventually(cf.Cf("restage", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
					Eventually(helpers.CurlingAppRoot(appName)).Should(Equal(OK_RESPONSE))
				})

				It("has its public image cached in the private registry", func() {
					AssertImageAvailable(registryAddress, getAppImageName(appName))
				})
			})
		})
	})
})
