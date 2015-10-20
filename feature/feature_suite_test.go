package feature

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/cloudfoundry-incubator/docker-cache-acceptance-tests/commons"
)

var (
	context        helpers.SuiteContext
	startedApp     string
	toBeStoppedApp string
	stoppedApp     string
)

func TestApplications(t *testing.T) {
	RegisterFailHandler(Fail)

	SetDefaultEventuallyTimeout(2 * time.Minute)
	SetDefaultEventuallyPollingInterval(time.Second)

	config := helpers.LoadConfig()
	context = helpers.NewContext(config)
	environment := helpers.NewEnvironment(context)

	BeforeSuite(func() {
		environment.Setup()
		AssertDockerEnabled()

		startedApp = generator.RandomName()
		toBeStoppedApp = generator.RandomName()
		stoppedApp = generator.RandomName()

		Eventually(cf.Cf("docker-push", startedApp, "cloudfoundry/diego-docker-app:latest", "--no-start")).Should(Exit(0))
		Eventually(cf.Cf("start", startedApp), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
		Eventually(helpers.CurlingAppRoot(startedApp)).Should(Equal(OK_RESPONSE))

		Eventually(cf.Cf("docker-push", toBeStoppedApp, "cloudfoundry/diego-docker-app:latest", "--no-start")).Should(Exit(0))
		Eventually(cf.Cf("start", toBeStoppedApp), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
		Eventually(helpers.CurlingAppRoot(toBeStoppedApp)).Should(Equal(OK_RESPONSE))

		Eventually(cf.Cf("docker-push", stoppedApp, "cloudfoundry/diego-docker-app:latest", "--no-start")).Should(Exit(0))
		Consistently(helpers.CurlingAppRoot(stoppedApp)).Should(ContainSubstring(NOT_FOUND))
	})

	AfterSuite(func() {
		Eventually(cf.Cf("delete", "-r", "-f", startedApp)).Should(Exit(0))
		Eventually(cf.Cf("delete", "-r", "-f", toBeStoppedApp)).Should(Exit(0))
		Eventually(cf.Cf("delete", "-r", "-f", stoppedApp)).Should(Exit(0))

		EnableDockerFeatureFlag(context)

		environment.Teardown()
	})

	componentName := "Diego-Docker-Feature"

	rs := []Reporter{}

	if config.ArtifactsDirectory != "" {
		helpers.EnableCFTrace(config, componentName)
		rs = append(rs, helpers.NewJUnitReporter(config, componentName))
	}

	RunSpecsWithDefaultAndCustomReporters(t, componentName, rs)
}
