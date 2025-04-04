// A generated module for DaggerModuleDemo functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"dagger/dagger-module-demo/internal/dagger"
	"fmt"
	"log"
	"sync"
)

type DaggerModuleDemo struct{}

// Returns a container that echoes whatever string argument is provided
func (m *DaggerModuleDemo) ContainerEcho(stringArg string) *dagger.Container {
	return dag.Container().From("alpine:latest").WithExec([]string{"echo", stringArg})
}

// Returns lines that match a pattern in the files of the provided Directory
func (m *DaggerModuleDemo) GrepDir(ctx context.Context, directoryArg *dagger.Directory, pattern string) (string, error) {
	return dag.Container().
		From("alpine:latest").
		WithMountedDirectory("/mnt", directoryArg).
		WithWorkdir("/mnt").
		WithExec([]string{"grep", "-R", pattern, "."}).
		Stdout(ctx)
}

// CreateCluster spins up a k3s single-node cluster, deploys a simple pod, and polls for pod readiness.
func (m *DaggerModuleDemo) CreateCluster(ctx context.Context) (string, error) {
	// Create a container from the Rancher k3s image.
	// We run the k3s server with a token but wrap it in a shell command
	// using the 'timeout' utility so that it stops after 10 seconds.
	cluster := dag.Container().
		From("k0sproject/k0s:latest").
		WithExec([]string{
			"sh", "-c", "timeout 10 k0s server --single",
		})

	// Force the container to run and capture its stdout.
	output, err := cluster.Stdout(ctx)
	if err != nil {
		return "", err
	}
	return output, nil

}

// MultiContainerResult holds the outputs from three containers and a collated string.
type MultiContainerResult struct {
	Output1        string `json:"output1"`
	Output2        string `json:"output2"`
	Output3        string `json:"output3"`
	CollatedOutput string `json:"collated_output"`
}

// RunThreeContainers demonstrates running three containers concurrently and collating their outputs.
func (m *DaggerModuleDemo) RunThreeContainers(ctx context.Context) (MultiContainerResult, error) {
	// Create three containers from the alpine image that simply echo a message.
	container1 := dag.Container().
		From("k0sproject/k0s:latest")
	container2 := dag.Container().
		From("rancher/k3s:latest")
	container3 := dag.Container().
		From("rancher/k3s:latest")

	var wg sync.WaitGroup
	wg.Add(3)

	var out1, out2, out3 string
	var err1, err2, err3 error

	// Run container1 concurrently.
	go func() {
		defer wg.Done()
		out1, err1 = container1.WithExec([]string{
			"sh", "-c", "timeout 10 k0s server --single",
		}).
			Stdout(ctx)
		if err1 != nil {
			log.Printf("Error from container1: %v", err1)
		}
	}()

	// Run container2 concurrently.
	go func() {
		defer wg.Done()
		out2, err2 = container2.WithExec([]string{
			"sh", "-c", "timeout 10 /bin/k3s server --token mysecret2",
		}).
			Stdout(ctx)
		if err2 != nil {
			log.Printf("Error from container2: %v", err2)
		}
	}()

	// Run container3 concurrently.
	go func() {
		defer wg.Done()
		out3, err3 = container3.WithExec([]string{
			"sh", "-c", "timeout 10 /bin/k3s server --token mysecret3",
		}).
			Stdout(ctx)
		if err3 != nil {
			log.Printf("Error from container3: %v", err3)
		}
	}()

	// Wait for all containers to finish.
	wg.Wait()

	// Check if any container encountered an error.
	if err1 != nil {
		return MultiContainerResult{}, err1
	}
	if err2 != nil {
		return MultiContainerResult{}, err2
	}
	if err3 != nil {
		return MultiContainerResult{}, err3
	}

	// Collate outputs from all containers.
	collated := fmt.Sprintf("Container1: %s\nContainer2: %s\nContainer3: %s\n", out1, out2, out3)

	// Build and return the composite result.
	result := MultiContainerResult{
		Output1:        out1,
		Output2:        out2,
		Output3:        out3,
		CollatedOutput: collated,
	}
	return result, nil
}

// ListGitHubRepos takes a GitHub token (as an ordered argument) and uses the GitHub CLI
// to list repositories. The token is passed in via an environment variable.
func (m *DaggerModuleDemo) ListGithubRepos(ctx context.Context) (string, error) {
	// This function builds a container that installs the GitHub CLI (gh) on an Ubuntu image.
	// It downloads the linux_arm64 tarball for gh version 2.14.3, extracts it,
	// copies the gh binary into /usr/local/bin, and finally runs "gh repo list".
	secret := dag.SetSecret("gh_token", "GITHUB_TOKEN")
	return dag.Container().
		From("ubuntu:latest").
		// Set the GitHub token as an environment variable so that gh can use it.
		WithSecretVariable("GH_TOKEN", secret).
		WithExec([]string{"sh", "-c", `
apt-get update && apt-get install -y curl tar && \
curl -LO https://github.com/cli/cli/releases/download/v2.14.3/gh_2.14.3_linux_arm64.tar.gz && \
tar -xzf gh_2.14.3_linux_arm64.tar.gz && \
cp gh_2.14.3_linux_arm64/bin/gh /usr/local/bin/ && \
gh repo list
`}).
		Stdout(ctx)
}

func (m *DaggerModuleDemo) ParseBase64JSON(ctx context.Context) (string, error) {
	// Hardcoded base64 encoded JSON.
	// The JSON is: {"VAR1": "value1", "VAR2": "value2", "VAR3": "value3"}
	encoded := "eyJWQVJBIjogInZhbHVlMSIsICJWQVJBIjogInZhbHVlMiIsICJWQVJBIjogInZhbHVlMyJ9"

	return dag.Container().
		From("alpine:latest").
		WithEnvVariable("ENCODED_JSON", encoded).
		WithExec([]string{"sh", "-c", `
apk add --no-cache jq && \
echo "$ENCODED_JSON" | base64 -d | jq -r '.VAR1, .VAR2, .VAR3'
`}).
		Stdout(ctx)
}
