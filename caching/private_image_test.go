package caching

import (
	"fmt"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/cloudfoundry-incubator/docker-cache-acceptance-tests/commons"

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
		spaceGuid := GuidForSpaceName(context.RegularUserContext().Space)
		config := helpers.LoadConfig()
		payload := fmt.Sprintf(createDockerAppPayload,
			appName,
			spaceGuid,
			config.DockerPrivateImage,
			config.DockerUser,
			config.DockerPassword,
			config.DockerEmail,
		)
		Eventually(cf.Cf("curl", "/v2/apps", "-X", "POST", "-d", payload)).Should(Exit(0))
		domain := helpers.LoadConfig().AppsDomain
		Eventually(cf.Cf("create-route", context.RegularUserContext().Space, domain, "-n", appName)).Should(Exit(0))
		Eventually(cf.Cf("map-route", appName, domain, "-n", appName)).Should(Exit(0))
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
			Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "true")).Should(Exit(0))
			Eventually(cf.Cf("start", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
		})

		It("starts successfully", func() {
			Eventually(helpers.CurlingAppRoot(appName)).Should(Equal(OK_RESPONSE))
		})
	})

	Context("with caching disabled", func() {

		BeforeEach(func() {
			appName = generator.RandomName()
		})

		JustBeforeEach(func() {
			Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "false")).Should(Exit(0))
			Eventually(cf.Cf("start", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(1))

			appLogs := GetAppLogs(appName)
			Expect(appLogs).To(ContainSubstring("failed to fetch metadata"))
		})

		Context("and then restaged with caching enabled", func() {
			JustBeforeEach(func() {
				Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "true")).Should(Exit(0))
				Eventually(cf.Cf("restage", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
				Eventually(helpers.CurlingAppRoot(appName)).Should(Equal(OK_RESPONSE))
			})

			It("has its public image cached in the private registry", func() {
				AssertImageAvailable(helpers.LoadConfig().DockerRegistryAddress, getAppImageName(appName))
			})
		})
	})
})
