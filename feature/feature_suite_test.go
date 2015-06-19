package feature

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/cloudfoundry-incubator/docker-registry-acceptance-tests/commons"
)

var (
	context        helpers.SuiteContext
	startedApp     string
	toBeStoppedApp string
	stoppedApp     string
)

func TestApplications(t *testing.T) {
	RegisterFailHandler(Fail)

	SetDefaultEventuallyTimeout(time.Minute)
	SetDefaultEventuallyPollingInterval(time.Second)

	config := helpers.LoadConfig()
	context = helpers.NewContext(config)
	environment := helpers.NewEnvironment(context)

	BeforeSuite(func() {
		environment.Setup()
		AssertDockerEnabled()

		spaceGuid := GuidForSpaceName(context.RegularUserContext().Space)
		startedApp = generator.RandomName()
		toBeStoppedApp = generator.RandomName()
		stoppedApp = generator.RandomName()

		CreateDockerApp(context, startedApp, fmt.Sprintf(DOCKER_APP_PAYLOAD_TEMPLATE, startedApp, spaceGuid, DIEGO_DOCKER_APP_IMAGE))
		Eventually(cf.Cf("start", startedApp), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
		Eventually(helpers.CurlingAppRoot(startedApp)).Should(Equal(OK_RESPONSE))

		CreateDockerApp(context, toBeStoppedApp, fmt.Sprintf(DOCKER_APP_PAYLOAD_TEMPLATE, toBeStoppedApp, spaceGuid, DIEGO_DOCKER_APP_IMAGE))
		Eventually(cf.Cf("start", toBeStoppedApp), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
		Eventually(helpers.CurlingAppRoot(toBeStoppedApp)).Should(Equal(OK_RESPONSE))

		CreateDockerApp(context, stoppedApp, fmt.Sprintf(DOCKER_APP_PAYLOAD_TEMPLATE, stoppedApp, spaceGuid, DIEGO_DOCKER_APP_IMAGE))
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
