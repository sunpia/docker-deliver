package main

import (
	"context"
	"log"
	"os/exec"
)

func main() {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx,
		"docker", "buildx", "bake",
		"--file", "C:\\Users\\darmh\\OneDrive\\Documents\\0_project\\docker-deliver\\example\\docker-compose.base.yaml",

		"--metadata-file", `C:\Users\darmh\AppData\Local\Temp\compose-build-metadataFile-687197424.json`,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}
	log.Printf("Command output:\n%s", output)
}
