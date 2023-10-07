package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Kind represents the KinD module for Dagger.
type Kind struct{}

// Exec executes the kind command with the given arguments. If no arguments are given, --help is used.
func (m *Kind) Exec(_ context.Context, args []string) *Container {
	// If no arguments are given, we'll default to --help.
	if len(args) == 0 {
		args = []string{"--help"}
	}

	return container().kind(args)
}

// Cluster returns the cluster with the given name. If no name is given, the default name, kind, is used. If a cluster
// already exists with the given name, it marks the cluster as existing to avoid creating it again.
func (m *Kind) Cluster(ctx context.Context, name string) (*Cluster, error) {
	if name == "" {
		name = "kind"
	}

	clusters, err := m.Exec(ctx, []string{"get", "clusters"}).Stdout(ctx)
	if err != nil {
		return nil, err
	}

	exist := false

	for _, cluster := range strings.Split(clusters, "\n") {
		if cluster == name {
			fmt.Printf("Cluster %s already exists.\n", name)
			exist = true
			break
		}
	}

	network := ""

	if exist {
		network, err = getContainersNetwork(ctx, fmt.Sprintf("^%s-control-plane-*", name))
		if err != nil {
			return nil, err
		}

		engineNetwork, err := getContainersNetwork(ctx, "^dagger-engine-*")
		if err != nil {
			return nil, err
		}

		// If the cluster is not on the same network as the engine containers, we'll return an error to avoid using the
		// cluster. This is important to be able to access the cluster from other containers using the IP address of
		// the cluster. If the cluster is not on the same network, the IP address of the cluster won't be accessible
		// from other containers.
		if network != engineNetwork {
			return nil, fmt.Errorf("cluster %s is already created on a different network. Please delete the cluster and try again", name)
		}
	}

	return &Cluster{Name: name, Network: network, Exists: exist}, nil
}

// Cluster represents a KinD cluster.
type Cluster struct {
	Name    string // Name is the name of the cluster.
	Network string // Network is the network name of the cluster.
	Exists  bool   // Exists is true if the cluster exists on the host with the given name.
}

// Create creates the cluster if it doesn't already exist.
func (m *Cluster) Create(ctx context.Context) (*Cluster, error) {
	if m.Exists {
		fmt.Printf("Cluster %s already exists.\n", m.Name)
		return m, nil
	}

	// Get the network name for the engine containers to ensure the cluster is created on the same network. It's
	// important to use the same network to be able to access the cluster from other containers using the IP address of
	// the cluster.
	network, err := getContainersNetwork(ctx, "^dagger-engine-*")
	if err != nil {
		return m, err
	}

	_, err = container().
		WithEnvVariable("KIND_EXPERIMENTAL_DOCKER_NETWORK", network).
		kind([]string{"create", "cluster", "--name", m.Name}).Sync(ctx)
	if err != nil {
		return m, err
	}

	m.Network = network
	m.Exists = true

	return m, nil
}

// Kubeconfig returns the kubeconfig file for the cluster. If internal is true, the internal address is used. Otherwise,
// the external address is used in the kubeconfig. Internal config is useful for running kubectl commands from within
// other containers.
func (m *Cluster) Kubeconfig(ctx context.Context, internal bool) (*File, error) {
	cmd := []string{"export", "kubeconfig", "--name", m.Name}

	if internal {
		cmd = append(cmd, "--internal")
	}

	file := container().kind(cmd).Directory("/root/.kube").File("config")

	if internal {
		ip, err := getClusterIPAddress(ctx, m.Network, m.Name)
		if err != nil {
			return nil, err
		}

		contents, err := file.Contents(ctx)
		if err != nil {
			return nil, err
		}

		contents = strings.ReplaceAll(contents, fmt.Sprintf("https://%s-control-plane:6443", m.Name), fmt.Sprintf("https://%s:6443", ip))

		file = dag.Directory().WithNewFile("config", contents).File("config")
	}

	return file, nil
}

// Logs returns the directory containing the cluster logs.
func (m *Cluster) Logs(_ context.Context) *Directory {
	dir := filepath.Join("tmp", m.Name, "logs")

	return container().kind([]string{"export", "logs", dir, "--name", m.Name}).Directory(dir)
}

// Delete deletes the cluster if it exists.
func (m *Cluster) Delete(ctx context.Context) (*Cluster, error) {
	if !m.Exists {
		fmt.Printf("Cluster %s doesn't exist.\n", m.Name)
		return m, nil
	}

	_, err := container().kind([]string{"delete", "cluster", "--name", m.Name}).Sync(ctx)
	if err != nil {
		return m, err
	}

	m.Network = ""
	m.Exists = false

	return m, nil
}

// exec executes the kind command with the given arguments.
func (c *Container) kind(args []string) *Container {
	return c.WithExec(append([]string{"kind"}, args...))
}

// container returns a container with the docker and kind binaries installed and the docker socket mounted. As last
// step, it adds a CACHE_BUSTER environment variable to the container to avoid using the cache when running the
// commands.
func container() *Container {
	socket := "/var/run/docker.sock"

	if env := os.Getenv("DOCKER_HOST"); env != "" {
		socket = strings.TrimPrefix(env, "unix://")
	}

	return dag.Container().
		From("alpine:latest").
		WithUser("root").
		WithExec([]string{"apk", "add", "--no-cache", "docker", "kind"}).
		WithUnixSocket("/var/run/docker.sock", dag.Host().UnixSocket(socket)).
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano))
}

// getContainersNetwork returns the network name of the given container.
func getContainersNetwork(ctx context.Context, name string) (string, error) {
	out, err := container().
		WithExec([]string{"docker", "container", "ls", "--filter", fmt.Sprintf("name=%s", name), "--format", "{{.Networks}}", "-n", "1"}).
		Stdout(ctx)

	return strings.TrimSpace(out), err
}

// getClusterIPAddress returns the IP address of the cluster control plane node. This is useful to access the cluster
// from other containers in the same network.
func getClusterIPAddress(ctx context.Context, network, name string) (string, error) {
	out, err := container().
		WithExec([]string{"docker", "inspect", fmt.Sprintf("%s-control-plane", name), "--format", fmt.Sprintf("{{.NetworkSettings.Networks.%s.IPAddress}}", network)}).
		Stdout(ctx)

	return strings.TrimSpace(out), err
}
