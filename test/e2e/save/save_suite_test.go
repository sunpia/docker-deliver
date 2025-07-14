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

	ginkgo.It("should save images and generate docker-compose file", func() {
		outputDir := "tmp"
		_ = os.RemoveAll(outputDir)
		err := os.MkdirAll(outputDir, 0755)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to create output directory")

		// Get git hash for tag
		gitHashCmd := exec.Command("git", "rev-parse", "--short", "HEAD")
		gitHashBytes, err := gitHashCmd.Output()
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to get git hash")
		gitHash := string(gitHashBytes)
		gitHash = string([]byte(gitHash)[:len(gitHash)-1]) // remove trailing newline

		// Run the docker-deliver save command
		cmd := exec.Command(
			"docker-deliver", "save",
			"-f", "example/docker-compose.base.yaml",
			"-f", "example/docker-compose.extend.yaml",
			"-o", outputDir,
			"-t", gitHash,
		)
		output, err := cmd.CombinedOutput()
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to execute docker-deliver save: %s", string(output))

		// Check for images.tar
		imagesTarPath := outputDir + "/images.tar"
		info, err := os.Stat(imagesTarPath)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "images.tar should exist in output dir")
		gomega.Expect(info.Size()).Should(gomega.BeNumerically(">", 0), "images.tar should not be empty")

		// Check for docker-compose.generated.yaml
		composePath := outputDir + "/docker-compose.generated.yaml"
		info, err = os.Stat(composePath)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "docker-compose.generated.yaml should exist in output dir")
		gomega.Expect(info.Size()).Should(gomega.BeNumerically(">", 0), "docker-compose.generated.yaml should not be empty")

		// Run docker compose down to clean up any previous resources
		downCmd := exec.Command(
			"docker", "compose",
			"-f", composePath,
			"down", "--rmi", "all",
		)
		downOutput, err := downCmd.CombinedOutput()
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to run docker compose down: %s", string(downOutput))

		upCmd := exec.Command(
			"docker", "compose",
			"-f", composePath,
			"up", "-d",
		)
		upOutput, err := upCmd.CombinedOutput()
		gomega.Expect(err).To(gomega.HaveOccurred(), "docker compose up should fail before images are loaded: %s", string(upOutput))

		// Load images.tar using appropriate docker load command based on OS
		var loadCmd *exec.Cmd
		if os.Getenv("OS") == "Windows_NT" {
			loadCmd = exec.Command("docker", "load", "-i", imagesTarPath)
		} else {
			loadCmd = exec.Command("sh", "-c", "docker load < "+imagesTarPath)
		}
		loadOutput, err := loadCmd.CombinedOutput()
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to load images: %s", string(loadOutput))
		gomega.Expect(string(loadOutput)).To(gomega.ContainSubstring("Loaded image"), "docker load output should mention loaded image")
		upCmd = exec.Command(
			"docker", "compose",
			"-f", composePath,
			"up", "-d",
		)
		upOutput, err = upCmd.CombinedOutput()
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "docker compose up should succeed after images are loaded: %s", string(upOutput))

		downCmd = exec.Command(
			"docker", "compose",
			"-f", composePath,
			"down", "--rmi", "all",
		)
		downOutput, err = downCmd.CombinedOutput()
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "failed to clean up resources: %s", string(downOutput))
		_ = os.RemoveAll(outputDir)
	})
})
