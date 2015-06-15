package caching

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
	. "github.com/cloudfoundry-incubator/docker-registry-acceptance-tests/commons"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Docker Registry", func() {
	type consulServiceInfo struct {
		Address string
	}

	var (
		registryAddress string
		imageAddress    string
		imageName       string
	)

	config := helpers.LoadConfig()

	runDockerCommand := func(args ...string) {
		args = append(config.DockerParameters, args...)

		cmd := exec.Command(config.DockerExecutable, args...)
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		err := cmd.Run()

		Expect(err).ShouldNot(HaveOccurred(), fmt.Sprintf("Error executing [%s %s]", config.DockerExecutable, strings.Join(args, " ")))
	}

	BeforeEach(func() {
		imageName = generator.RandomName()
		registryAddress = config.DockerRegistryAddress
		imageAddress = fmt.Sprintf("%s/%s", registryAddress, imageName)

		runDockerCommand("pull", "busybox")
		runDockerCommand("tag", "busybox", imageAddress)
	})

	Describe("Docker Registry", func() {
		It("accepts push requests", func() {
			runDockerCommand("push", imageAddress)
		})

		It("can be searched for images", func() {
			runDockerCommand("push", imageAddress)
			AssertImageAvailable(registryAddress, imageName)
		})

		It("accepts pull requests", func() {
			runDockerCommand("push", imageAddress)
			// Clean the local copy
			runDockerCommand("rmi", imageAddress)

			// Make sure we can pull it from private registry
			runDockerCommand("pull", imageAddress)
		})
	})
})
