package metadata

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
	})

	AfterSuite(func() {
		EnableDockerFeatureFlag(context)

		environment.Teardown()
	})

	componentName := "Diego-Docker-Metadata"

	rs := []Reporter{}

	if config.ArtifactsDirectory != "" {
		helpers.EnableCFTrace(config, componentName)
		rs = append(rs, helpers.NewJUnitReporter(config, componentName))
	}

	RunSpecsWithDefaultAndCustomReporters(t, componentName, rs)
}
