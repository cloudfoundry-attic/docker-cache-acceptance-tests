package feature

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/cloudfoundry-incubator/docker-cache-acceptance-tests/commons"
	"github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

type appEntity struct {
	Application app `json:"entity"`
}

type app struct {
	Instances int `json:"instances"`
	Memory    int `json:"memory"`
}

type domainsResult struct {
	Domains []domainEntity `json:"resources"`
}

type domainEntity struct {
	Domain domain `json:"entity"`
}

type domain struct {
	Name string `json:"name"`
}

type routesResult struct {
	Routes []routeEntity `json:"resources"`
}

type routeEntity struct {
	Route route `json:"entity"`
}

type route struct {
	Host string `json:"host"`
}

type envResult struct {
	Env env `json:"environment_json"`
}

type env struct {
	Key string
}

var _ = Describe("Diego Docker Support", func() {

	Context("with some Docker Apps already running", func() {

		Context("and Docker disabled", func() {
			BeforeEach(func() {
				DisableDockerFeatureFlag(context)

				By("ensure docker apps are eventually stopped")
				Eventually(helpers.CurlingAppRoot(startedApp), 1*time.Minute).Should(ContainSubstring(NOT_FOUND))
				Consistently(helpers.CurlingAppRoot(startedApp)).Should(ContainSubstring(NOT_FOUND))

				Eventually(helpers.CurlingAppRoot(toBeStoppedApp), 1*time.Minute).Should(ContainSubstring(NOT_FOUND))
				Consistently(helpers.CurlingAppRoot(toBeStoppedApp)).Should(ContainSubstring(NOT_FOUND))

				Consistently(helpers.CurlingAppRoot(stoppedApp)).Should(ContainSubstring(NOT_FOUND))
			})

			Context("when app start is requested", func() {
				var session *Session

				BeforeEach(func() {
					session = cf.Cf("start", stoppedApp)
				})

				It("should fail", func() {
					Expect(session.Wait()).To(Exit(1))
					Expect(session.Out).To(Say("Docker support has not been enabled"))
				})
			})

			Context("when app stop is requested", func() {
				It("should succeed", func() {
					Eventually(cf.Cf("stop", toBeStoppedApp)).Should(Exit(0))
				})
			})

			Context("when app is restaged", func() {
				var session *Session

				BeforeEach(func() {
					session = cf.Cf("restage", startedApp)
				})

				It("should fail", func() {
					Expect(session.Wait()).To(Exit(1))
					Expect(session.Out).To(Say("Docker support has not been enabled"))
				})
			})

			Context("when app is scaled horizontally", func() {
				BeforeEach(func() {
					Eventually(cf.Cf("scale", startedApp, "-i", "3")).Should(Exit(0))

					session := cf.Cf("curl", fmt.Sprintf("/v2/apps/%s", GuidForAppName(startedApp)))
					Expect(session.Wait()).To(Exit(0))

					response := appEntity{}
					err := json.Unmarshal(session.Out.Contents(), &response)
					Expect(err).ToNot(HaveOccurred())
					Expect(response.Application.Instances).To(Equal(3))
				})

				AfterEach(func() {
					Eventually(cf.Cf("scale", startedApp, "-i", "1")).Should(Exit(0))
				})

				It("shoud not start the app", func() {
					Consistently(helpers.CurlingAppRoot(startedApp)).Should(ContainSubstring(NOT_FOUND))
				})
			})

			Context("when app is scaled vertically", func() {
				BeforeEach(func() {
					Eventually(cf.Cf("curl", fmt.Sprintf("/v2/apps/%s", GuidForAppName(startedApp)), "-X", "PUT", "-d", `{"memory":128}`)).Should(Exit(0))

					session := cf.Cf("curl", fmt.Sprintf("/v2/apps/%s", GuidForAppName(startedApp)))
					Expect(session.Wait()).To(Exit(0))

					response := appEntity{}
					err := json.Unmarshal(session.Out.Contents(), &response)
					Expect(err).ToNot(HaveOccurred())
					Expect(response.Application.Memory).To(Equal(128))
				})

				It("shoud not start the app", func() {
					Consistently(helpers.CurlingAppRoot(startedApp)).Should(ContainSubstring(NOT_FOUND))
				})
			})

			Context("when a route is mapped", func() {
				var (
					domain    string
					routeName string
				)

				listRoutes := func() []routeEntity {
					routesQuery := cf.Cf("curl", fmt.Sprintf("/v2/apps/%s/routes", GuidForAppName(startedApp)))
					Expect(routesQuery.Wait()).To(Exit(0))

					routesQueryResult := routesResult{}
					err := json.Unmarshal(routesQuery.Out.Contents(), &routesQueryResult)
					Expect(err).ToNot(HaveOccurred())

					return routesQueryResult.Routes
				}

				BeforeEach(func() {
					uuid, err := uuid.NewV4()
					Expect(err).NotTo(HaveOccurred())
					routeName = uuid.String()

					sharedDomainsQuery := cf.Cf("curl", "/v2/shared_domains")
					Eventually(sharedDomainsQuery.Wait()).Should(Exit(0))

					domainsResult := domainsResult{}
					err = json.Unmarshal(sharedDomainsQuery.Out.Contents(), &domainsResult)
					Expect(err).ToNot(HaveOccurred())

					domain = domainsResult.Domains[0].Domain.Name

					Eventually(cf.Cf("map-route", startedApp, domain, "-n", routeName)).Should(Exit(0))

					Expect(len(listRoutes())).To(Equal(2))
				})

				It("shoud not start the app", func() {
					Consistently(helpers.CurlingAppRoot(startedApp)).Should(ContainSubstring(NOT_FOUND))
				})

				Context("and then unmapped", func() {
					BeforeEach(func() {
						Eventually(cf.Cf("unmap-route", startedApp, domain, "-n", routeName)).Should(Exit(0))

						Expect(len(listRoutes())).To(Equal(1))
					})

					It("shoud not start the app", func() {
						Consistently(helpers.CurlingAppRoot(startedApp)).Should(ContainSubstring(NOT_FOUND))
					})
				})
			})

			Context("when env is set", func() {
				BeforeEach(func() {
					Eventually(cf.Cf("set-env", startedApp, "Key", "Value")).Should(Exit(0))

					envQuery := cf.Cf("curl", fmt.Sprintf("/v2/apps/%s/env", GuidForAppName(startedApp)))
					Expect(envQuery.Wait()).To(Exit(0))

					envQueryResult := envResult{}
					err := json.Unmarshal(envQuery.Out.Contents(), &envQueryResult)
					Expect(err).ToNot(HaveOccurred())

					Expect(envQueryResult.Env.Key).To(Equal("Value"))
				})

				It("shoud not start the app", func() {
					Consistently(helpers.CurlingAppRoot(startedApp)).Should(ContainSubstring(NOT_FOUND))
				})
			})

			Context("and then enabled back", func() {
				BeforeEach(func() {
					EnableDockerFeatureFlag(context)
				})

				It("should bring desired apps back up", func() {
					Eventually(helpers.CurlingAppRoot(startedApp), 2*time.Minute).Should(Equal(OK_RESPONSE))
					Consistently(helpers.CurlingAppRoot(startedApp)).Should(Equal(OK_RESPONSE))

					Consistently(helpers.CurlingAppRoot(stoppedApp)).Should(ContainSubstring(NOT_FOUND))
				})
			})

		})
	})

})
