package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const (
	DefaultClusterName = "kind"
	DefaultDockerHost  = "unix:///var/run/docker.sock"
)

// Kind represents the KinD module for Dagger.type
type Kind struct{}

// Cli returns a container with the kind binary installed.
func (m *Kind) Cli(
	ctx context.Context,
	// docker host (default: unix:///var/run/docker.sock)
	dockerHost Optional[string],
) (*Container, error) {
	dockerHostVal := dockerHost.GetOr(DefaultDockerHost)

	// Get the network name for the engine containers to ensure the cluster is created on the same network. It's
	// important to use the same network to be able to access the cluster from other containers using the IP address of
	// the cluster.
	network, err := getContainersNetwork(ctx, "^dagger-engine-*", dockerHostVal)
	if err != nil {
		return nil, err
	}

	return container(dockerHostVal).WithEnvVariable("KIND_EXPERIMENTAL_DOCKER_NETWORK", network), nil
}

// Connect returns a container with the kubeconfig file mounted to be able to access the given cluster. If the cluster
// doesn't exist, it returns an error.
func (m *Kind) Connect(
	ctx context.Context,
	// name of the cluster
	name Optional[string],
	// docker host (default: unix:///var/run/docker.sock)
	dockerHost Optional[string],
) (*Container, error) {
	dockerHostVal := dockerHost.GetOr(DefaultDockerHost)

	cluster, err := m.Cluster(ctx, name, dockerHost)
	if err != nil {
		return nil, err
	}

	if !cluster.Exists {
		return nil, fmt.Errorf("cluster %s doesn't exist", cluster.Name)
	}

	kubeconfig, err := cluster.Kubeconfig(ctx, Opt[bool](true))
	if err != nil {
		return nil, err
	}

	return container(dockerHostVal).
		WithMountedFile("/root/.kube/config", kubeconfig).
		WithEnvVariable("KUBECONFIG", "/root/.kube/config"), nil
}

// Cluster returns the cluster with the given name. If no name is given, the default name, kind, is used. If a cluster
// already exists with the given name, it marks the cluster as existing to avoid creating it again.
func (m *Kind) Cluster(
	ctx context.Context,
	// name of the cluster
	name Optional[string],
	// docker host (default: unix:///var/run/docker.sock)
	dockerHost Optional[string],
) (*Cluster, error) {

	var (
		clusterName   = name.GetOr(DefaultClusterName)
		dockerHostVal = dockerHost.GetOr(DefaultDockerHost)
	)

	clusters, err := kind(container(dockerHostVal), []string{"get", "clusters"}).Stdout(ctx)
	if err != nil {
		return nil, err
	}

	exist := false

	for _, cluster := range strings.Split(clusters, "\n") {
		println(cluster)
		if cluster == clusterName {
			fmt.Printf("Cluster %s already exists.\n", clusterName)
			exist = true
			break
		}
	}

	network := ""

	if exist {
		network, err = getContainersNetwork(ctx, fmt.Sprintf("^%s-control-plane-*", clusterName), dockerHostVal)
		if err != nil {
			return nil, err
		}

		engineNetwork, err := getContainersNetwork(ctx, "^dagger-engine-*", dockerHostVal)
		if err != nil {
			return nil, err
		}

		// If the cluster is not on the same network as the engine containers, we'll return an error to avoid using the
		// cluster. This is important to be able to access the cluster from other containers using the IP address of
		// the cluster. If the cluster is not on the same network, the IP address of the cluster won't be accessible
		// from other containers.
		if network != engineNetwork {
			return nil, fmt.Errorf("cluster %s is already created on a different network. Please delete the cluster and try again", clusterName)
		}
	}

	return &Cluster{
		Name:       clusterName,
		Network:    network,
		Exists:     exist,
		DockerHost: dockerHostVal,
	}, nil
}

// Cluster represents a KinD cluster.
type Cluster struct {
	Name       string // Name is the name of the cluster.
	Network    string // Network is the network name of the cluster.
	Exists     bool   // Exists is true if the cluster exists on the host with the given name.
	DockerHost string // DockerHost is the docker host of the host.
}

// Create creates the cluster if it doesn't already exist.
func (m *Cluster) Create(ctx context.Context) (string, error) {
	if m.Exists {
		return fmt.Sprintf("cluster %s already exists", m.Name), nil
	}

	// Get the network name for the engine containers to ensure the cluster is created on the same network. It's
	// important to use the same network to be able to access the cluster from other containers using the IP address of
	// the cluster.
	network, err := getContainersNetwork(ctx, "^dagger-engine-*", m.DockerHost)
	if err != nil {
		return "", err
	}

	container := container(m.DockerHost).WithEnvVariable("KIND_EXPERIMENTAL_DOCKER_NETWORK", network)

	args := []string{"create", "cluster"}

	if m.Name != DefaultClusterName {
		args = append(args, "--name", m.Name)
	}

	_, err = kind(container, args).Sync(ctx)
	if err != nil {
		return "", err
	}

	m.Network = network
	m.Exists = true

	return fmt.Sprintf("cluster %s created", m.Name), nil
}

// Kubeconfig returns the kubeconfig file for the cluster. If internal is true, the internal address is used. Otherwise,
// the external address is used in the kubeconfig. Internal config is useful for running kubectl commands from within
// other containers.
func (m *Cluster) Kubeconfig(
	ctx context.Context,
	// internal is true if the internal address should be used in the kubeconfig.
	internal Optional[bool],
) (*File, error) {
	internalVal := internal.GetOr(false)

	cmd := []string{"export", "kubeconfig"}

	if m.Name != DefaultClusterName {
		cmd = append(cmd, "--name", m.Name)
	}

	if internalVal {
		cmd = append(cmd, "--internal")
	}

	file := kind(container(m.DockerHost), cmd).Directory("/root/.kube").File("config")

	if internalVal {
		ip, err := getClusterIPAddress(ctx, m.Network, m.Name, m.DockerHost)
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

	args := []string{"export", "logs", dir}

	if m.Name != DefaultClusterName {
		args = append(args, "--name", m.Name)
	}

	return kind(container(m.DockerHost), args).Directory(dir)
}

// Delete deletes the cluster if it exists.
func (m *Cluster) Delete(ctx context.Context) (string, error) {
	if !m.Exists {
		return fmt.Sprintf("cluster %s doesn't exist", m.Name), nil
	}

	args := []string{"delete", "cluster"}

	if m.Name != DefaultClusterName {
		args = append(args, "--name", m.Name)
	}

	_, err := kind(container(m.DockerHost), args).Sync(ctx)
	if err != nil {
		return "", err
	}

	m.Network = ""
	m.Exists = false

	return fmt.Sprintf("cluster %s deleted", m.Name), nil
}

// exec executes the kind command with the given arguments.
func kind(c *Container, args []string) *Container {
	return c.WithExec(append([]string{"kind"}, args...), ContainerWithExecOpts{ExperimentalPrivilegedNesting: true})
}

// container returns a container with the docker and kind binaries installed and the docker socket mounted. As last
// step, it adds a CACHE_BUSTER environment variable to the container to avoid using the cache when running the
// commands.
func container(dockerHost string) *Container {
	container := dag.Container().
		From("alpine/k8s:1.28.3").
		WithUser("root").
		WithWorkdir("/").
		WithExec([]string{"apk", "add", "--no-cache", "docker", "kind", "k9s"})

	if dockerHost != "" {
		switch {
		case strings.HasPrefix(dockerHost, "unix://"):
			dockerHost = strings.TrimPrefix(dockerHost, "unix://")

			container = container.WithUnixSocket("/var/run/docker.sock", dag.Host().UnixSocket(dockerHost))
			container = container.WithEnvVariable("DOCKER_HOST", "unix:///var/run/docker.sock")
		case strings.HasPrefix(dockerHost, "tcp://"):
			container = container.WithEnvVariable("DOCKER_HOST", dockerHost)
		}
	}

	return container.WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano))
}

// getContainersNetwork returns the network name of the given container.
func getContainersNetwork(ctx context.Context, name, dockerHost string) (string, error) {
	out, err := container(dockerHost).
		WithExec([]string{"docker", "container", "ls", "--filter", fmt.Sprintf("name=%s", name), "--format", "{{.Networks}}", "-n", "1"}).
		Stdout(ctx)

	return strings.TrimSpace(out), err
}

// getClusterIPAddress returns the IP address of the cluster control plane node. This is useful to access the cluster
// from other containers in the same network.
func getClusterIPAddress(ctx context.Context, network, name, dockerHost string) (string, error) {
	out, err := container(dockerHost).
		WithExec([]string{"docker", "inspect", fmt.Sprintf("%s-control-plane", name), "--format", fmt.Sprintf("{{.NetworkSettings.Networks.%s.IPAddress}}", network)}).
		Stdout(ctx)

	return strings.TrimSpace(out), err
}
