package save_e2e

import (
	"os"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/sunpia/docker-deliver/test/e2e"
)

func TestSave(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Save Suite")
}

var _ = Describe("Compile Pkg", func() {

	BeforeEach(func() {
		// Jump to the project root directory
		projectRoot, err := ProjectRootPath()
		Expect(err).NotTo(HaveOccurred())
		err = os.Chdir(projectRoot)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should install the CLI binary", func() {
		InstallApplication()
		cmd := exec.Command("docker-deliver", "--help")
		output, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), "failed to execute 'docker-deliver --help'")
		Expect(string(output)).To(ContainSubstring("Usage"), "output should contain 'Usage'")
	})

})
