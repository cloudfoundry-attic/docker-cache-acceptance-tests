package docker

import (
	"fmt"
	"regexp"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Docker Registry", func() {
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

	JustBeforeEach(func() {
		spaceGuid := guidForSpaceName(context.RegularUserContext().Space)
		payload := fmt.Sprintf(createDockerAppPayload, appName, spaceGuid)

		createDockerApp(appName, payload)

		Eventually(cf.Cf("set-env", appName, "DIEGO_DOCKER_CACHE", "true"))
		Eventually(cf.Cf("start", appName), DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT).Should(Exit(0))
		Eventually(helpers.CurlingAppRoot(appName)).Should(Equal("0"))
	})

	AfterEach(func() {
		Eventually(cf.Cf("logs", appName, "--recent")).Should(Exit())
		Eventually(cf.Cf("delete", appName, "-f")).Should(Exit(0))
	})

	Describe("running the app with private registry", func() {
		var imageName string
		var address string

		JustBeforeEach(func() {
			cfLogs := cf.Cf("logs", appName, "--recent")
			Expect(cfLogs.Wait()).To(Exit(0))
			contents := string(cfLogs.Out.Contents())

			//TODO: Replace with list all droplets API (/v3/droplets)
			r := regexp.MustCompile(".*Docker image will be cached as ([0-z.:]+)/([0-z-]+)")
			imageParts := r.FindStringSubmatch(contents)
			Expect(len(imageParts)).Should(Equal(3))

			address = imageParts[1]
			imageName = imageParts[2]
		})

		It("stores the public image in the private registry", func() {
			assertImageAvailable(address, imageName)
		})
	})
})
