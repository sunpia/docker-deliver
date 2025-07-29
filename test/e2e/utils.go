package e2e

import (
	"context"
	"fmt"
	"os/exec"
)

// InstallApplication runs 'make install' to build the application.
func InstallApplication(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "make", "install")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install application: %w", err)
	}
	return nil
}

func BuildApplication(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "make", "build")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build application: %w", err)
	}
	return nil
}

func ProjectRootPath(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get project root path: %w", err)
	}
	return string(output[:len(output)-1]), nil // remove trailing newline
}
