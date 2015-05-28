package docker

import (
	"fmt"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("A Docker App", func() {
	var appName string
	var createDockerAppPayload string

	BeforeEach(func() {
		appName = generator.RandomName()

		createDockerAppPayload = `{
			"name": "%s",
			"memory": 512,
			"instances": 1,
			"disk_quota": 1024,
			"space_guid": "%s",
			"docker_image": "cloudfoundry/diego-docker-app:latest",
			"command": "/myapp/dockerapp",
			"diego": true
		}`
	})

	Context("pushed to Diego", func() {
		JustBeforeEach(func() {
			spaceGuid := guidForSpaceName(context.RegularUserContext().Space)
			payload := fmt.Sprintf(createDockerAppPayload, appName, spaceGuid)

			createDockerApp(appName, payload)
		})

		AfterEach(func() {
			Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
			Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
		})

		Context("with caching enabled", func() {

			JustBeforeEach(func() {
				Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "true"))
				Eventually(cf.Cf("start", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
				Eventually(helpers.CurlingAppRoot(appName)).Should(Equal("0"))
			})

			It("has its public image cached in the private registry", func() {
				assertImageAvailable(getAppImageDetails(appName))
			})
		})

		Context("with caching disabled", func() {

			JustBeforeEach(func() {
				Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "false"))
				Eventually(cf.Cf("start", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
				Eventually(helpers.CurlingAppRoot(appName)).Should(Equal("0"))
			})

			Context("and then restaged with caching enabled", func() {

				JustBeforeEach(func() {
					Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "true"))
					Eventually(cf.Cf("restage", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
					Eventually(helpers.CurlingAppRoot(appName)).Should(Equal("0"))
				})

				It("has its public image cached in the private registry", func() {
					assertImageAvailable(getAppImageDetails(appName))
				})
			})
		})
	})
})
