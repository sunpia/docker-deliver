package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/client"
)

func createTarFromDir(dir string) (io.ReadCloser, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	err := filepath.Walk(dir, func(file string, fi fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories (they are included implicitly)
		if fi.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dir, file)
		if err != nil {
			return err
		}

		// Open file
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		// Prepare header
		header, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath) // use forward slashes for tar format

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// Copy file contents
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func main() {
	composePath := "./example/test-compose.yaml"
	workingDir := "./example"
	if len(os.Args) > 1 {
		composePath = os.Args[1]
	}

	ctx := context.Background()
	project, err := cli.ProjectFromOptions(ctx, &cli.ProjectOptions{
		ConfigPaths: []string{composePath},
		WorkingDir:  workingDir,
		Environment: map[string]string{},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading compose file: %v\n", err)
		os.Exit(1)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Docker client: %v\n", err)
		os.Exit(1)
	}
	defer cli.Close()

	for _, svc := range project.Services {
		if svc.Build != nil {
			fmt.Printf("Building image for service: %s\n", svc.Name)
			buildCtx, err := createTarFromDir(svc.Build.Context)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening build context: %v\n", err)
				continue
			}
			buildOptions := build.ImageBuildOptions{
				BuildArgs:  svc.Build.Args,
				Dockerfile: svc.Build.Dockerfile,
				Target:     svc.Build.Target,
			}

			defer buildCtx.Close()
			resp, err := cli.ImageBuild(ctx, buildCtx, buildOptions)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to build image for %s: %v\n", svc.Name, err)
				continue
			}
			defer resp.Body.Close()
			// Optionally, print build output
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				var msg struct {
					Stream string `json:"stream"`
				}
				line := scanner.Text()
				if err := json.Unmarshal([]byte(line), &msg); err != nil {
					fmt.Println("Error:", err)
					continue
				}
				fmt.Print(msg.Stream)
			}
		}
	}
}
