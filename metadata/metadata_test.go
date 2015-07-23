package metadata

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/cloudfoundry-incubator/docker-cache-acceptance-tests/commons"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Diego Docker Metadata", func() {
	var appName string

	curlingFunc := func(appName, path string) func() string {
		return func() string {
			return helpers.CurlApp(appName, path)
		}
	}

	Context("with application that has custom docker build instructions", func() {
		BeforeEach(func() {
			appName = generator.RandomName()

			Eventually(cf.Cf("docker-push", appName, "cloudfoundry/diego-docker-app-custom:latest", "--no-start")).Should(Exit(0))
		})

		AfterEach(func() {
			Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
			Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
		})

		It("starts", func() {
			Eventually(cf.Cf("start", appName)).Should(Exit(0))
			Consistently(helpers.CurlingAppRoot(appName)).Should(ContainSubstring(OK_RESPONSE))
		})

		It("listens on custom port", func() {
			Eventually(cf.Cf("start", appName)).Should(Exit(0))
			Consistently(curlingFunc(appName, "/env")).Should(ContainSubstring(`"PORT":"7070"`))
			Consistently(curlingFunc(appName, "/env")).Should(MatchRegexp(`"CF_INSTANCE_PORTS":"\d+:7070`))
		})

		It("uses a custom user", func() {
			Eventually(cf.Cf("start", appName)).Should(Exit(0))
			Consistently(curlingFunc(appName, "/env")).Should(ContainSubstring(`"USER":"dockeruser"`))
		})
	})

})
