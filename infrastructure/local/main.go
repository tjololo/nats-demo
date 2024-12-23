package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const ProjectRootEnvVar = "NATS_DEMO_PROJECT_ROOT"
const BuildFolder = "bin"

func main() {
	args := os.Args[1:]
	command := "start"
	if len(args) > 0 {
		command = args[0]
	}
	switch command {
	case "start":
		if !isFloxEnvActive() {
			fmt.Println("************** Warning **************")
			fmt.Println("* Flox environment is not active")
			fmt.Println("* It is recommended to activate flox environment to ensure dependencies are installed.")
			fmt.Println("* Install guides can be found at https://flox.dev/docs/install-flox/")
			fmt.Println("* To enable flox environment run: flox activate")
			fmt.Println("************************************")
		}
		checkPrerequisites()
		startKindCluster()
		tarballPath := buildServiceImage()
		defer os.Remove(tarballPath)
		loadImageIndorCluster(tarballPath)
		deployNatsHelmChart()
		deployNatsDemoManifests()
		fmt.Println("Nats demo is up and running")
		fmt.Println("To access publisher web api run: `kubectl port-forward svc/publisher 8181:8181`")
		fmt.Println("To view logs run: `kubectl logs -f -l app=subscriber`")
	case "stop":
		fmt.Println("Stopping kind cluster")
		stopKindCluster()
	default:
		fmt.Println("Unknown argument")
		fmt.Println("Valid arguments are: start, stop")
		os.Exit(1)
	}
}

func isFloxEnvActive() bool {
	fmt.Println("Checking if flox environment is active")
	cmd, err := exec.Command("flox", "envs", "--active", "--json").Output()
	if err != nil {
		fmt.Println("Failed to check if flox environment is active")
		os.Exit(1)
	}
	var envs []FloxEnv
	err = json.Unmarshal(cmd, &envs)
	if err != nil {
		fmt.Println("Failed to parse active flox environment")
		os.Exit(1)
	}
	if len(envs) == 0 {
		fmt.Println("No active flox environment")
		return false
	}
	for _, env := range envs {
		if env.Pointer.Name == "nats-demo" {
			return true
		}
	}
	return false
}

func checkPrerequisites() {
	fmt.Println("Checking prerequisites")
	cmd := exec.Command("which", "kind")
	err := cmd.Run()
	if err != nil {
		fmt.Println("Kind is not installed")
		os.Exit(1)
	}
	cmd = exec.Command("which", "ko")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Ko is not installed")
		os.Exit(1)
	}
	cmd = exec.Command("which", "helm")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Helm is not installed")
		os.Exit(1)
	}
	cmd = exec.Command("which", "kubectl")
	err = cmd.Run()
	if err != nil {
		fmt.Println("Kubectl is not installed")
		os.Exit(1)
	}
	if os.Getenv(ProjectRootEnvVar) == "" {
		fmt.Println(ProjectRootEnvVar + " environment variable is not set")
		os.Exit(1)
	}
}

func startKindCluster() {
	if isNatDemoClusterRunning() {
		fmt.Println("Kind cluster `nats-demo` is already running")
	} else {
		fmt.Println("Starting kind cluster")
		cmd := exec.Command("kind", "create", "cluster", "--name", "nats-demo")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println("Failed to start kind cluster")
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Kind cluster started")
	}
}

func stopKindCluster() {
	if !isNatDemoClusterRunning() {
		fmt.Println("Kind cluster `nats-demo` is not running")
	} else {
		fmt.Println("Stopping kind cluster")
		cmd := exec.Command("kind", "delete", "cluster", "--name", "nats-demo")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			fmt.Println("Failed to stop kind cluster")
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Kind cluster stopped")
	}
}

func isNatDemoClusterRunning() bool {
	fmt.Println("Getting running kind clusters")
	cmd := exec.Command("kind", "get", "clusters")
	var out bytes.Buffer
	cmd.Stdin = os.Stdin
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Failed to get running kind clusters")
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Running kind clusters: " + out.String())
	clusters := strings.TrimSpace(out.String())
	if clusters != "" {
		clustersSlice := strings.Split(clusters, "\n")
		for _, cluster := range clustersSlice {
			if cluster == "nats-demo" {
				return true
			}
		}
	}
	return false
}

func buildServiceImage() string {
	fmt.Println("Building service image with ko")
	root := os.Getenv(ProjectRootEnvVar)
	if root == "" {
		fmt.Println(ProjectRootEnvVar + " environment variable is not set")
		os.Exit(1)
	}
	tmpDir, err := os.MkdirTemp("", "nats-demo")
	if err != nil {
		fmt.Println("Failed to create temporary directory")
		fmt.Println(err)
		os.Exit(1)
	}
	tarball := filepath.Join(tmpDir, "service.tar")
	args := []string{"build", "github.com/tjololo/nats-demo/service", "--base-import-paths", "-t=local", "--tarball=" + tarball, "--push=false"}
	cmd := exec.Command("ko", args...)
	cmd.Dir = path.Join(root, "service")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Failed to build service image")
		fmt.Println(stderr.String())
		os.Exit(1)
	}
	fmt.Println("Service image built")
	fmt.Println(out.String())
	return tarball
}

func loadImageIndorCluster(tarballPath string) {
	fmt.Println("Loading image into kind cluster")
	cmd := exec.Command("kind", "load", "image-archive", tarballPath, "--name", "nats-demo")
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Failed to load image into kind cluster")
		fmt.Println(stdErr.String())
		os.Exit(1)
	}
	fmt.Println("Image loaded into kind cluster")
}

func deployNatsHelmChart() {
	fmt.Println("Deploying nats helm chart")
	cmd := exec.Command("helm", "repo", "add", "nats", "https://nats-io.github.io/k8s/helm/charts")
	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut
	var stdErrAdd bytes.Buffer
	cmd.Stderr = &stdErrAdd
	err := cmd.Run()
	if err != nil {
		fmt.Println("Failed to add nats helm chart repository")
		fmt.Println(stdErrAdd.String())
		os.Exit(1)
	}
	cmd = exec.Command("helm", "upgrade", "--install", "nats", "nats/nats", "--set", "natsBox.enabled=false")
	stdOut.Reset()
	cmd.Stdout = &stdOut
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Failed to deploy nats helm chart")
		fmt.Println(stdErr.String())
		os.Exit(1)
	}
	fmt.Println("Nats helm chart deployed")
}

func deployNatsDemoManifests() {
	fmt.Println("Deploying nats demo services")
	cmd := exec.Command("kubectl", "apply", "-f", "manifests")
	root := os.Getenv(ProjectRootEnvVar)
	if root == "" {
		fmt.Println(ProjectRootEnvVar + " environment variable is not set")
		os.Exit(1)
	}
	cmd.Dir = path.Join(root, "infrastructure", "local")
	var stdOut bytes.Buffer
	cmd.Stdout = &stdOut
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Failed to deploy nats demo services")
		fmt.Println(stdErr.String())
		os.Exit(1)
	}
	fmt.Println("Nats publisher and subscriber deployed")
}

func getPodmanSocketInfoIfAwailable() (string, bool) {
	var podmanInfos []PodmanMachineInfo
	cmd := exec.Command("podman", "machine", "inspect")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Failed to get podman machine info")
		return "", false
	}
	err = json.Unmarshal(out, &podmanInfos)
	if err != nil {
		fmt.Println("Failed to parse podman machine info")
		return "", false
	}
	for _, podmanInfo := range podmanInfos {
		if podmanInfo.State == "running" && podmanInfo.ConnectionInfo.PodmanSocket.Path != "" {
			return fmt.Sprintf("unix:///%s", podmanInfo.ConnectionInfo.PodmanSocket.Path), true
		}
	}
	return "", false
}

type FloxEnv struct {
	Path    string         `json:"path"`
	Pointer FloxEnvPointer `json:"pointer"`
	Type    string         `json:"type"`
}

type FloxEnvPointer struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

type PodmanMachineInfo struct {
	State          string `json:"State"`
	ConnectionInfo struct {
		PodmanSocket struct {
			Path string `json:"Path"`
		} `json:"PodmanSocket"`
	} `json:"ConnectionInfo"`
}
