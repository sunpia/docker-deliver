package e2e

import (
	"fmt"
	"os/exec"
)

// InsatllApplication runs 'make install' to build the application.
func InstallApplication() error {
	cmd := exec.Command("make", "install")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install application: %w", err)
	}
	return nil
}

func BuildApplication() error {
	cmd := exec.Command("make", "build")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build application: %w", err)
	}
	return nil
}

func ProjectRootPath() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get project root path: %w", err)
	}
	return string(output[:len(output)-1]), nil // remove trailing newline
}
