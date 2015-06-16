package smoke

import (
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Diego Docker Feature Flag", func() {
	It("is enabled", func() {
		session := cf.Cf("feature-flag", "diego_docker")
		Eventually(session.Wait()).Should(Exit(0))

		Expect(session.Out).To(Say("enabled"))
	})
})
