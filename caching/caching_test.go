package caching

import (
	"fmt"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/cloudfoundry-incubator/docker-registry-acceptance-tests/commons"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("A Docker App", func() {
	var appName string

	BeforeEach(func() {
		appName = generator.RandomName()
	})

	Context("pushed to Diego", func() {
		JustBeforeEach(func() {
			spaceGuid := GuidForSpaceName(context.RegularUserContext().Space)
			payload := fmt.Sprintf(DOCKER_APP_PAYLOAD_TEMPLATE, appName, spaceGuid, DIEGO_DOCKER_APP_IMAGE)

			CreateDockerApp(context, appName, payload)
		})

		AfterEach(func() {
			Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
			Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
		})

		Context("with caching enabled", func() {

			JustBeforeEach(func() {
				Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "true"))
				Eventually(cf.Cf("start", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
				Eventually(helpers.CurlingAppRoot(appName)).Should(Equal(OK_RESPONSE))
			})

			It("has its public image cached in the private registry", func() {
				AssertImageAvailable(getAppImageDetails(appName))
			})
		})

		Context("with caching disabled", func() {

			JustBeforeEach(func() {
				Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "false"))
				Eventually(cf.Cf("start", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
				Eventually(helpers.CurlingAppRoot(appName)).Should(Equal(OK_RESPONSE))
			})

			Context("and then restaged with caching enabled", func() {

				JustBeforeEach(func() {
					Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "true"))
					Eventually(cf.Cf("restage", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
					Eventually(helpers.CurlingAppRoot(appName)).Should(Equal(OK_RESPONSE))
				})

				It("has its public image cached in the private registry", func() {
					AssertImageAvailable(getAppImageDetails(appName))
				})
			})
		})
	})
})