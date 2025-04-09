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
	"bufio"
	"context"

	"dagger/dagger-module-demo/internal/dagger"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

// DaggerModuleDemo is our module type with a CacheVolume field.
type DaggerModuleDemo struct {
	CacheVolume *dagger.CacheVolume
}

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

// ParseBase64JSON demonstrates internal logging and capturing logs from two containers.
// The function returns a Base64 encoded JSON string that combines logs from container1 and container2.
func (m *DaggerModuleDemo) LoggingTest(ctx context.Context) (string, error) {
	// Log the start of the function execution to stderr.
	log.Printf("Starting LoggingTest function execution")

	// Run the first container: for demonstration, we use an Alpine image that echoes a simple log message.
	container1 := dag.Container().From("alpine:latest").WithExec([]string{"echo", "container1 log"})
	// Run the second container similarly.
	container2 := dag.Container().From("alpine:latest").WithExec([]string{"echo", "container2 log"})

	// Retrieve the output from container1.
	out1, err := container1.Stdout(ctx)
	if err != nil {
		log.Printf("Error retrieving output from container1: %v", err)
		return "", err
	}
	log.Printf("Container1 log: %s", out1)

	// Retrieve the output from container2.
	out2, err := container2.Stdout(ctx)
	if err != nil {
		log.Printf("Error retrieving output from container2: %v", err)
		return "", err
	}
	log.Printf("Container2 log: %s", out2)

	// Combine logs from both containers into a map.
	combinedLogs := map[string]string{
		"container1": out1,
		"container2": out2,
	}

	// Marshal the combined logs map to JSON.
	jsonData, err := json.Marshal(combinedLogs)
	if err != nil {
		log.Printf("Error marshalling combined logs: %v", err)
		return "", err
	}

	// Optionally, encode the JSON data in Base64 to ensure a safe string return.
	// encoded := base64.StdEncoding.EncodeToString(jsonData)
	log.Printf("Returning final encoded output.")

	// Return the Base64 encoded string.
	return string(jsonData), nil
}

// ParseMultipleLogs runs a container that outputs multiple log lines.
// The function logs internal messages (sent to stderr) and returns the container's stdout.
func (m *DaggerModuleDemo) ParseMultipleLogs(ctx context.Context) (string, error) {
	// Log the start of the function execution.
	log.Printf("Starting ParseMultipleLogs: container will produce 101 log statements over ~15 seconds")

	// Get the current timestamp.
	now := time.Now().Format(time.RFC3339)
	// Create a container from Alpine that runs a shell loop.
	// The shell command prints a log statement 101 times, sleeping 150ms between each.
	container := dag.Container().
		From("alpine:latest").
		WithExec([]string{"sh", "-c", "for i in $(seq 1 101); do echo \"Container! log statement " + now + " $i\"; sleep 0.15; done"})

	// Retrieve the container's stdout, which will include all log statements
	_, err := container.Stdout(ctx)
	if err != nil {
		log.Printf("Error retrieving stdout from container: %v", err)
		return "", err
	}

	log.Printf("Container produced logs successfully")

	// Return the container's stdout. The Dagger CLI will serialize this return value as JSON on stdout.
	return "done", nil
}

// ParseSingleLog runs a container that outputs a single log message.
// The function logs diagnostic information and returns the container's stdout.
func (m *DaggerModuleDemo) ParseSingleLog(ctx context.Context) (string, error) {
	// Log the start of execution.
	log.Printf("Starting ParseSingleLog function execution")

	// Create a container from Alpine that outputs a single log message.
	container := dag.Container().
		From("alpine:latest").
		WithExec([]string{"echo", "Single log message"})

	// Retrieve the stdout from the container.
	stdout, err := container.Stdout(ctx)
	if err != nil {
		log.Printf("Error retrieving stdout from container: %v", err)
		return "", err
	}

	// Log the container's output to stderr.
	log.Printf("Container produced single log: %s", stdout)

	// Return the stdout output.
	return stdout, nil
}

func (m *DaggerModuleDemo) RunContainerWithLogging(ctx context.Context) (string, error) {
	// Log the start of the function execution (goes to stderr).
	fmt.Println("Starting RunContainerWithLogging function execution")

	// Create a volume to store container logs.
	logVolume := dag.CacheVolume("container-log-volume")

	// Define a shell command that writes log statements to /logs/container.log.
	// This command loops 10 times with a 1-second sleep between iterations.
	date := time.Now().Format(time.RFC3339)
	shellCommand := `for i in $(seq 1 10); do echo "Container log statement $i at ` + date + `" >> /logs/container.log; sleep 1; done`

	// Create a container from Alpine Linux.
	// Mount the volume at /logs so the container can write its log file there.
	container := dag.Container().
		From("alpine:latest").
		WithMountedCache("/logs", logVolume).
		WithExec([]string{"sh", "-c", shellCommand})

	// Run the container.
	// We don't care about its stdout here because the logs are written to /logs/container.log.
	_, err := container.Stdout(ctx)
	if err != nil {
		log.Printf("Error running container command: %v", err)
		return "", err
	}

	// Retrieve the log file content from the volume.
	// We use the same container to read from the mounted directory.
	// logs, err := container.Directory("/logs").File("container.log").Contents(ctx)
	// if err != nil {
	// 	log.Printf("Error retrieving container log file: %v", err)
	// 	return "", err
	// }

	// Generate metadata.
	timestamp := time.Now().Format(time.RFC3339)

	// Create a fancy summary using ANSI escape codes.
	// Colors: Blue title, green for timestamp, yellow for the dagger.
	fancySummary := fmt.Sprintf("\033[1;34mModule Run Summary\033[0m\n"+
		"\033[1;32mTimestamp:\033[0m %s\n"+
		"\033[1;33mDagger:\033[0m 🗡️\n\n"+
		"\033[1;36mContainer Log Output:\033[0m\n%s",
		timestamp, "logs")

	// Log the fancy summary to stderr.
	fmt.Printf("Fancy Summary Generated:\n%s", fancySummary)

	// Return the fancy summary.
	// The Dagger CLI will serialize this return value into JSON and write it to stdout.
	return fancySummary, nil
}

func (m *DaggerModuleDemo) LoggingExample(ctx context.Context) (string, error) {
	// Using fmt to print to stdout.
	fmt.Println("fmt.Println: This message is printed to stdout.")
	fmt.Printf("fmt.Printf: Here's a formatted message with a value: %v\n", 42)

	// Using log to print to stderr.
	log.Println("log.Println: This message is printed to stderr.")
	log.Printf("log.Printf: Here's a formatted message on stderr: %v", "hello stderr")

	// Simulate an error using fmt.Errorf.
	// This creates an error value that we log, but we don't return it as the function error.
	errExample := fmt.Errorf("fmt.Errorf: This is an example error")
	log.Printf("Logging an error with log.Printf: %v", errExample)

	// Using an alternative method: writing directly to stderr.
	// (Normally, you'd use log or a dedicated logging framework.)
	// os.Stderr.WriteString("Directly writing to stderr using os.Stderr.WriteString\n")
	// Uncomment the above line if you want to test direct writes.

	// Final summary: This is the value that will be returned and captured by the Dagger CLI.
	summary := "LoggingExample executed. Check your terminal for differences between stdout and stderr logs."
	return summary, nil
}

// InteractiveTerminal demonstrates a basic interactive terminal in a Dagger function.
// It prints a simple menu, reads user input, and then uses the selection to run a container
// that echoes the chosen message.
// DOESN'T WORK
func (m *DaggerModuleDemo) InteractiveTerminal(ctx context.Context) (string, error) {
	// Print the interactive menu to stdout.
	fmt.Println("=== Interactive Terminal ===")
	fmt.Println("Please select an option:")
	fmt.Println("1) Say Hello")
	fmt.Println("2) Echo a custom message")
	fmt.Print("Enter your choice (1 or 2): ")

	// Read input from the user.
	reader := bufio.NewReader(os.Stdin)
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading input: %v", err)
	}
	choice = strings.TrimSpace(choice)

	// Determine the message based on the selection.
	var message string
	switch choice {
	case "1":
		message = "Hello, World!"
	case "2":
		fmt.Print("Enter your custom message: ")
		custom, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("error reading custom message: %v", err)
		}
		message = strings.TrimSpace(custom)
	default:
		fmt.Println("Invalid selection, defaulting to 'Hello, World!'")
		message = "Hello, World!"
	}

	// Use a container that echoes the selected message.
	container := dag.Container().From("alpine:latest").
		WithExec([]string{"echo", message})

	// Get the output from the container.
	output, err := container.Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("error running echo container: %v", err)
	}

	// Display the container's output.
	fmt.Println("Container output:", output)

	// Return the output (this will be serialized by the Dagger CLI).
	return output, nil
}

// DeployAndPollK8s demonstrates asynchronous orchestration in a Dagger function.
// It starts a k0s Kubernetes cluster as a service, polls for its readiness asynchronously,
// deploys a simple pod, and then polls the cluster for that pod.
func (m *DaggerModuleDemo) DeployAndPollK8s(ctx context.Context) (*dagger.Container, error) {
	// 1. Start the k0s container as a service.
	//    We assume that "k0sproject/k0s:latest" is a valid image that runs a k0s server.
	k0sContainer := dag.Container().
		From("k0sproject/k0s:latest").
		WithExec([]string{"k0s", "server"})
	// Convert the container to a service.

	// // 2. Poll for k0s readiness asynchronously.
	// readyCh := make(chan struct{})
	// go func() {
	// 	// In a real-world scenario, you would poll the k0s API or check logs.
	// 	// Here we simulate waiting for readiness by sleeping 15 seconds.
	// 	fmt.Println("Polling for k0s readiness...")
	// 	time.Sleep(29 * time.Second)
	// 	fmt.Println("k0s is now ready.")
	// 	readyCh <- struct{}{}
	// }()

	// // Wait for readiness or timeout.
	// select {
	// case <-readyCh:
	// 	// Proceed once the service is ready.
	// case <-time.After(30 * time.Second):
	// 	return "", fmt.Errorf("k0s cluster did not become ready in time")
	// }

	// 3. Retrieve the kubeconfig from the k0s container.
	//    Assume k0s writes its admin kubeconfig to /var/lib/k0s/pki/admin.conf.
	kubeconfigContents, err := k0sContainer.Directory("/var/lib/k0s/pki").File("admin.conf").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("error retrieving kubeconfig: %v", err)
	}
	fmt.Println("Retrieved kubeconfig from k0s.")

	// Create a directory with the kubeconfig file for mounting.
	kubeconfigDir := dag.Directory().WithNewFile("admin.conf", kubeconfigContents)

	// 4. Create a simple deployment manifest (a pod running nginx).
	deploymentManifest := `
apiVersion: v1
kind: Pod
metadata:
  name: hello-world
spec:
  containers:
  - name: hello
    image: nginx
`
	manifestDir := dag.Directory().WithNewFile("deployment.yaml", deploymentManifest)

	k0sService := k0sContainer.AsService()
	// 5. Deploy the manifest using a kubectl container.
	dag.Container().
		From("bitnami/kubectl:latest").
		WithMountedDirectory("/", manifestDir).
		WithMountedDirectory("/kubeconfig", kubeconfigDir).
		WithEnvVariable("KUBECONFIG", "/kubeconfig/admin.conf").
		// Bind the k0s service so kubectl can reach the cluster.
		WithServiceBinding("k8s", k0sService).
		WithExec([]string{"kubectl", "apply", "-f", "deployment.yaml"})

	// deployerOutput, err := deployer.Stdout(ctx)
	// if err != nil {
	// 	return nil, fmt.Errorf("error deploying manifest: %v", err)
	// }
	// fmt.Println("Deployment manifest applied.")

	// 6. Start a poller container to check for the deployed pod.
	return dag.Container().
		From("bitnami/kubectl:latest").
		WithMountedDirectory("/kubeconfig", kubeconfigDir).
		WithEnvVariable("KUBECONFIG", "/kubeconfig/admin.conf").
		WithServiceBinding("k8s", k0sService).
		WithExec([]string{"sh", "-c", "for i in $(seq 1 10); do echo 'Polling pods...'; kubectl get pods; sleep 2; done"}), nil

	// pollerOutput, err := poller.Stdout(ctx)
	// if err != nil {
	// 	return "", fmt.Errorf("error polling cluster: %v", err)
	// }

	// // 7. Combine and return the outputs.
	// summary := fmt.Sprintf("Deployment output:\n%s\n\nPolling output:\n%s", deployerOutput, pollerOutput)
	// return summary, nil
}

// StartK0sCluster starts a k0s cluster and then blocks to keep it running.
// For demonstration purposes, it sleeps for 5 minutes before returning.
func (m *DaggerModuleDemo) StartK0sCluster(ctx context.Context) (string, error) {
	// Print a message to indicate we're starting the cluster.
	fmt.Println("Starting k0s cluster...")

	// Create the k0s container. We assume the k0s image runs a Kubernetes control plane
	// when invoked with "k0s server --disable-worker".
	k0sContainer := dag.Container().
		From("k0sproject/k0s:latest").
		WithExec([]string{"k0s", "server"})

	// if err != nil {
	// 	return "", fmt.Errorf("error while starting k0s: %v", err)
	// }
	// Convert the container to a service so that it runs in the background and can be referenced.
	k0sService := k0sContainer.AsService()
	var _ interface{}
	var err error

	_, err = k0sContainer.Directory("/var/lib/k0s/pki").File("admin.conf").Contents(ctx)
	if err != nil {
		return "", fmt.Errorf("error retrieving kubeconfig: %v", err)
	}
	fmt.Println("Retrieved kubeconfig from k0s.")

	// k0sContainerService.Start(ctx)

	// Log that the service has been started.
	fmt.Println("k0s cluster service started.")

	// For demonstration purposes, block for 5 minutes to keep the cluster running.
	// (In practice, the lifetime of the cluster is bound to the function's execution.)
	fmt.Println("Cluster will remain running for 5 minutes. You can connect to it from other containers in this function.")
	_, err = dag.Container().
		// From("alpine:latest").
		From("bitnami/kubectl:latest").
		WithServiceBinding("k8s", k0sService).
		WithExec([]string{"sleep", "300"}).Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("error while sleeping to keep cluster alive: %v", err)
	}

	// After the sleep period, return a summary message.
	return "k0s cluster was running for 5 minutes", nil
}

// DeployK3sAndApp deploys a k3s cluster, installs an app via Helm,
// tests the app via curl, and returns the output.
func (m *DaggerModuleDemo) DeployK3sAndApp(ctx context.Context) (*dagger.Container, error) {
	// Create a new k3s cluster instance named "test".
	cacheBuster := fmt.Sprintf("%d", time.Now().UnixNano())
	container := dag.Container().
		From("alpine:latest").
		WithEnvVariable("CACHE_BUSTER x", cacheBuster).
		WithExec([]string{"sh", "-c", "echo Cache buster is $CACHE_BUSTER && sleep 1"})

	_, err := container.Stdout(ctx)
	if err != nil {
		return nil, err
	}

	k3sCluster := dag.K3S("test " + fmt.Sprint(rand.Int()))
	// Get the k3s server container.
	// kServer := k3sCluster.Container().AsService()
	kServer := k3sCluster.Server()

	// Start the k3s cluster.
	kServer, err = kServer.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start k3s cluster: %w", err)
	}
	fmt.Println("k3s cluster started as a service.")

	// Retrieve the endpoint for the cluster's service on port 80 over HTTP.
	ep, err := kServer.Endpoint(ctx, dagger.ServiceEndpointOpts{
		Port:   80,
		Scheme: "http",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cluster endpoint: %w", err)
	}

	// Use an alpine/helm container to:
	// 1. Install kubectl (via apk add).
	// 2. Set KUBECONFIG from the k3s cluster's kubeconfig.
	// 3. Install the nginx chart via Helm.
	// 4. Run curl to test the deployment using the endpoint.
	// out, err := dag.Container().From("alpine/helm").
	return dag.Container().From("alpine/helm").
		WithExec([]string{"apk", "add", "kubectl"}).
		WithEnvVariable("KUBECONFIG", "/.kube/config").
		// Mount the kubeconfig file from the k3s cluster.
		WithFile("/.kube/config", k3sCluster.Config()).
		// Install the nginx chart.
		WithExec([]string{"helm", "install", "--wait", "--debug", "nginx", "oci://registry-1.docker.io/bitnamicharts/nginx"}).
		// Test the deployment using curl.
		WithExec([]string{"curl", "-sS", ep}), nil
	// Stdout(ctx)
}

// StartK3SAndOutputKubectlCommands starts a k3s cluster, retrieves the kubeconfig,
// and outputs a set of kubectl one-liners that you can use to interact with the cluster.
func (m *DaggerModuleDemo) StartK3SAndOutputKubectlCommands(ctx context.Context) (*dagger.Container, error) {
	// Create a new k3s cluster instance named "test".
	k3sCluster := dag.K3S("test")
	// Get the k3s server container.
	kServer := k3sCluster.Server()

	// Start the k3s cluster.
	// kServer, err := kServer.Start(ctx)
	_, err := kServer.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start k3s cluster: %w", err)
	}
	fmt.Println("k3s cluster started.")

	// Retrieve the kubeconfig as a dagger.File from the cluster.
	kubeconfigFile := k3sCluster.Config()
	// Convert the dagger.File to a string.
	kubeconfigContent, err := kubeconfigFile.Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve kubeconfig: %w", err)
	}

	// Prepare a set of kubectl one-liner commands.
	// These commands allow you to export the kubeconfig, check nodes and pods,
	// and optionally install applications or inspect pod logs.
	kubectlCommands := fmt.Sprintf(`# Export the kubeconfig for local kubectl usage:
echo '%s' > /tmp/k3s.yaml && export KUBECONFIG=/tmp/k3s.yaml

# Verify cluster nodes:
kubectl get nodes

# List all pods across all namespaces:
kubectl get pods --all-namespaces

# (Optional) Install an application using your manifest (replace <your-manifest.yaml> with your file):
kubectl apply -f <your-manifest.yaml>

# (Optional) Check logs for a pod (replace <pod-name> and <container-name> as needed):
kubectl logs <pod-name> -c <container-name>
`, kubeconfigContent)

	fmt.Print(kubectlCommands)

	// Return the one-liner commands so you can copy-paste them into your terminal.
	// return kubectlCommands, nil
	return k3sCluster.Container(), nil
}

// Init creates a cache volume, writes some initial files into it,
// attaches the volume to the module, and returns the module.
func (m *DaggerModuleDemo) Init(ctx context.Context) (*DaggerModuleDemo, error) {
	// Create a cache volume with a unique key.
	m.CacheVolume = dag.CacheVolume("chain-test-cache")

	// Run a container that mounts the cache volume at /data and writes two initial files.
	initContainer := dag.Container().
		From("alpine:latest").
		WithMountedCache("/data", m.CacheVolume).
		WithExec([]string{"sh", "-c", "mkdir -p /data && echo 'initial content for file1' > /data/file1.txt && echo 'initial content for file2' > /data/file2.txt"})

	// Run the container; we don’t need the output, just execute the commands.
	if _, err := initContainer.Stdout(ctx); err != nil {
		return nil, fmt.Errorf("Init: failed to write initial files: %w", err)
	}

	fmt.Println("Init: initial files written to cache volume.")
	return m, nil
}

// Continue runs a container that mounts the same cache volume at /data and writes
// an additional unique file into it. This allows multiple "continue" steps without conflicts.
func (m *DaggerModuleDemo) Continue(ctx context.Context) (*DaggerModuleDemo, error) {
	// Generate a unique identifier using the current timestamp.
	uniqueID := time.Now().UnixNano()
	fileName := fmt.Sprintf("/data/continue_%d.txt", uniqueID)
	fileContent := fmt.Sprintf("continued content at %d", uniqueID)

	// Run a container that mounts the cache volume and writes the unique file.
	continueContainer := dag.Container().
		From("alpine:latest").
		WithMountedCache("/data", m.CacheVolume).
		WithExec([]string{"sh", "-c", fmt.Sprintf("echo '%s' > %s", fileContent, fileName)})

	if _, err := continueContainer.Stdout(ctx); err != nil {
		return nil, fmt.Errorf("Continue: failed to write continued file: %w", err)
	}

	fmt.Printf("Continue: wrote unique file %s\n", fileName)
	return m, nil
}

func (m *DaggerModuleDemo) Output(ctx context.Context) (string, error) {
	// Run a container that mounts the cache volume at /data,
	// copies its contents to /output, and lists the /output directory.
	outputContainer := dag.Container().
		From("alpine:latest").
		WithMountedCache("/data", m.CacheVolume).
		WithExec([]string{"sh", "-c", "ls -a /data"})

	// Get the output from the container listing the files.
	output, err := outputContainer.Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("Output: failed to sync files to output: %w", err)
	}

	fmt.Println("Output: files synced to /output:")
	fmt.Println(output)
	return output, nil
}
