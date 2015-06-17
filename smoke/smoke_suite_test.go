package smoke

import (
	"regexp"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/cloudfoundry-incubator/docker-registry-acceptance-tests/commons"
)

var context helpers.SuiteContext

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
	})

	AfterSuite(func() {
		environment.Teardown()
	})

	componentName := "Diego-Docker-Smoke"

	rs := []Reporter{}

	if config.ArtifactsDirectory != "" {
		helpers.EnableCFTrace(config, componentName)
		rs = append(rs, helpers.NewJUnitReporter(config, componentName))
	}

	RunSpecsWithDefaultAndCustomReporters(t, componentName, rs)
}

func getAppImageDetails(appName string) (string, string) {
	contents := GetAppLogs(appName)

	//TODO: Replace with list all droplets API (/v3/droplets)
	r := regexp.MustCompile(".*Docker image will be cached as ([0-z.:]+)/([0-z-]+)")
	imageParts := r.FindStringSubmatch(contents)
	Expect(len(imageParts)).Should(Equal(3))

	return imageParts[1], imageParts[2]
}
