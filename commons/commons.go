package commons

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

const (
	CF_PUSH_TIMEOUT                       = 4 * time.Minute
	LONG_CURL_TIMEOUT                     = 4 * time.Minute
	DOCKER_IMAGE_DOWNLOAD_DEFAULT_TIMEOUT = 10 * time.Minute
	DOCKER_APP_PAYLOAD_TEMPLATE           = `{
			"name": "%s",
			"memory": 512,
			"instances": 1,
			"disk_quota": 1024,
			"space_guid": "%s",
			"docker_image": "cloudfoundry/diego-docker-app:latest",
			"command": "/myapp/dockerapp",
			"diego": true
		}`
	NOT_FOUND   = "404 Not Found"
	OK_RESPONSE = "0"
)

func GuidForAppName(appName string) string {
	cfApp := cf.Cf("app", appName, "--guid")
	Expect(cfApp.Wait()).To(Exit(0))

	appGuid := strings.TrimSpace(string(cfApp.Out.Contents()))
	Expect(appGuid).NotTo(Equal(""))
	return appGuid
}

func GuidForSpaceName(spaceName string) string {
	cfSpace := cf.Cf("space", spaceName, "--guid")
	Expect(cfSpace.Wait()).To(Exit(0))

	spaceGuid := strings.TrimSpace(string(cfSpace.Out.Contents()))
	Expect(spaceGuid).NotTo(Equal(""))
	return spaceGuid
}

func CreateDockerApp(context helpers.SuiteContext, appName, payload string) {
	Eventually(cf.Cf("curl", "/v2/apps", "-X", "POST", "-d", payload)).Should(Exit(0))
	domain := helpers.LoadConfig().AppsDomain
	Eventually(cf.Cf("create-route", context.RegularUserContext().Space, domain, "-n", appName)).Should(Exit(0))
	Eventually(cf.Cf("map-route", appName, domain, "-n", appName)).Should(Exit(0))
}

func AssertImageAvailable(registryAddress string, imageName string) {
	client := http.Client{}
	resp, err := client.Get(fmt.Sprintf("http://%s/v1/search?q=%s", registryAddress, imageName))
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	bytes, err := ioutil.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())
	Expect(string(bytes)).To(ContainSubstring("library/" + imageName))
}

func EnableDockerFeatureFlag(context helpers.SuiteContext) {
	cf.AsUser(context.AdminUserContext(), time.Minute, func() {
		Eventually(cf.Cf("enable-feature-flag", "diego_docker")).Should(Exit(0))
	})
}

func DisableDockerFeatureFlag(context helpers.SuiteContext) {
	cf.AsUser(context.AdminUserContext(), time.Minute, func() {
		Eventually(cf.Cf("disable-feature-flag", "diego_docker")).Should(Exit(0))
	})
}

func GetAppLogs(appName string) string {
	cfLogs := cf.Cf("logs", appName, "--recent")
	Expect(cfLogs.Wait()).To(Exit(0))
	return string(cfLogs.Out.Contents())
}

func AssertDockerEnabled() {
	session := cf.Cf("feature-flag", "diego_docker")
	Eventually(session.Wait()).Should(Exit(0))

	Expect(session.Out).To(Say("enabled"))
}
