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

var _ = Describe("Private Docker Image", func() {
	const createDockerAppPayload string = `{
			"name": "%s",
			"memory": 512,
			"instances": 1,
			"disk_quota": 1024,
			"space_guid": "%s",
			"docker_image": "%s",
			"docker_credentials_json" : {
				"docker_user" : "%s",
				"docker_password" : "%s",
				"docker_email" : "%s"
			},
			"command": "/myapp/dockerapp",
			"diego": true
		}`

	var appName string

	JustBeforeEach(func() {
		spaceGuid := guidForSpaceName(context.RegularUserContext().Space)
		config := helpers.LoadConfig()
		payload := fmt.Sprintf(createDockerAppPayload,
			appName,
			spaceGuid,
			config.DockerPrivateImage,
			config.DockerUser,
			config.DockerPassword,
			config.DockerEmail,
		)
		createDockerApp(appName, payload)
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

	Context("with caching enabled", func() {
		BeforeEach(func() {
			appName = generator.RandomName()
		})

		JustBeforeEach(func() {
			Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "true"))
			Eventually(cf.Cf("start", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
		})

		It("starts successfully", func() {
			Eventually(helpers.CurlingAppRoot(appName)).Should(Equal("0"))
		})
	})

	Context("with caching disabled", func() {
		BeforeEach(func() {
			appName = generator.RandomName()
		})

		JustBeforeEach(func() {
			Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "false"))
			Eventually(cf.Cf("start", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(1))

			appLogs := getAppLogs(appName)
			Expect(appLogs).To(ContainSubstring("failed to fetch metadata"))
		})

		Context("and then restaged with caching enabled", func() {
			JustBeforeEach(func() {
				Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "true"))
				Eventually(cf.Cf("restage", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
			})

			It("starts successfully", func() {
				Eventually(helpers.CurlingAppRoot(appName)).Should(Equal("0"))
			})

			It("has its public image cached in the private registry", func() {
				assertImageAvailable(getAppImageDetails(appName))
			})
		})
	})
})
