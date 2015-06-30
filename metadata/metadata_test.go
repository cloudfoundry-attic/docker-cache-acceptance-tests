package metadata

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

var _ = Describe("Diego Docker Metadata", func() {
	var appName string

	Context("with application that listens on a custom port", func() {
		BeforeEach(func() {
			appName = generator.RandomName()

			spaceGuid := GuidForSpaceName(context.RegularUserContext().Space)
			payload := fmt.Sprintf(DOCKER_APP_PAYLOAD_TEMPLATE, appName, spaceGuid, "cloudfoundry/diego-docker-app-custom:latest")

			CreateDockerApp(context, appName, payload)
		})

		AfterEach(func() {
			Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
			Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
		})

		It("starts", func() {
			Eventually(cf.Cf("start", appName)).Should(Exit(0))
			Consistently(helpers.CurlingAppRoot(appName)).Should(ContainSubstring(OK_RESPONSE))
		})
	})

})
