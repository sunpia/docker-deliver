package savee2e

import (
	"os"
	"os/exec"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sunpia/docker-deliver/test/e2e"
)

func TestSave(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Save Suite")
}

var _ = ginkgo.Describe("Compile Pkg", func() {

	ginkgo.BeforeEach(func() {
		// Jump to the project root directory
		projectRoot, err := e2e.ProjectRootPath()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		err = os.Chdir(projectRoot)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.It("should install the CLI binary", func() {
		_ = e2e.InstallApplication() // Ignoring error for test setup
		cmd := exec.Command("docker-deliver", "--help")
		output, err := cmd.CombinedOutput()
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to execute 'docker-deliver --help'")
		gomega.Expect(string(output)).To(gomega.ContainSubstring("Usage"), "output should contain 'Usage'")
	})

})
