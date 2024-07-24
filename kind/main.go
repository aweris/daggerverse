// A Dagger Module for integrating with the KinD
package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/samber/lo"

	"dagger/kind/internal/dagger"
)

func New(
	// Unix socket to connect to the external Docker Engine. Please carefully use this option it can expose your host to the container.
	//
	// +required
	socket *dagger.Socket,
) *Kind {
	return &Kind{DockerSocket: socket}
}

type Kind struct {
	// +private
	DockerSocket *dagger.Socket
}

// Container that contains the kind and k9s binaries
func (k *Kind) Container() *dagger.Container {
	return dag.Container().
		From("alpine/k8s:1.28.3").
		WithoutEntrypoint().
		WithUser("root").
		WithWorkdir("/").
		WithExec([]string{"apk", "add", "--no-cache", "docker", "kind", "k9s"}).
		WithUnixSocket("/var/run/docker.sock", k.DockerSocket).
		WithEnvVariable("DOCKER_HOST", "unix:///var/run/docker.sock")
}

func (k *Kind) Cluster(
	ctx context.Context,

	// Name of the cluster
	//
	// +optional
	// +default="kind"
	name string,
) (*Cluster, error) {

	exists, err := k.isClusterExist(context.Background(), name)
	if err != nil {
		return nil, err
	}

	// Get the network name for the engine containers to ensure the cluster is created on the same network. It's
	// important to use the same network to be able to access the cluster from other containers using the IP address of
	// the cluster.
	engineNetwork, err := k.getContainerNetwork(ctx, "^dagger-engine-*")
	if err != nil {
		return nil, err
	}

	if exists {
		clusterNetwork, err := k.getContainerNetwork(ctx, fmt.Sprintf("^%s-control-plane-*", name))
		if err != nil {
			return nil, err
		}

		// If the cluster is not on the same network as the engine containers, we'll return an error to avoid using the
		// cluster. This is important to be able to access the cluster from other containers using the IP address of
		// the cluster. If the cluster is not on the same network, the IP address of the cluster won't be accessible
		if engineNetwork != clusterNetwork {
			return nil, fmt.Errorf("cluster %s is not connected to the engine", name)
		}
	}

	return &Cluster{
		Name:    name,
		Exists:  exists,
		Network: engineNetwork,
		Kind:    k,
	}, nil
}

type Cluster struct {
	Name    string
	Network string
	Exists  bool

	// +private
	Kind *Kind
}

// Create creates the cluster if it doesn't already exist.
func (m *Cluster) Create(ctx context.Context) (string, error) {
	if m.Exists {
		return fmt.Sprintf("cluster %s already exists", m.Name), nil
	}

	args := []string{"kind", "create", "cluster"}

	if m.Name != "kind" {
		args = append(args, "--name", m.Name)
	}

	_, err := m.Kind.exec(ctx, m.Network, args...)
	if err != nil {
		return "", err
	}

	m.Exists = true

	return fmt.Sprintf("cluster %s created", m.Name), nil
}

// Exports cluster kubeconfig
func (m *Cluster) Kubeconfig(
	ctx context.Context,

	// If true, the internal address is used in the kubeconfig. This is useful for running kubectl commands from within other containers.
	//
	// +optional
	// +default=false
	internal bool,
) (*dagger.File, error) {
	cmd := []string{"kind", "export", "kubeconfig"}

	if m.Name != "kind" {
		cmd = append(cmd, "--name", m.Name)
	}

	if internal {
		cmd = append(cmd, "--internal")
	}

	ctr, err := m.Kind.exec(ctx, m.Network, cmd...)
	if err != nil {
		return nil, err
	}

	kubeconfig := ctr.Directory("/root/.kube").File("config")

	if internal {
		ip, err := m.Kind.getClusterIPAddress(ctx, m.Network, m.Name)
		if err != nil {
			return nil, err
		}

		contents, err := kubeconfig.Contents(ctx)
		if err != nil {
			return nil, err
		}

		contents = strings.ReplaceAll(contents, fmt.Sprintf("https://%s-control-plane:6443", m.Name), fmt.Sprintf("https://%s:6443", ip))

		kubeconfig = dag.Directory().WithNewFile("config", contents).File("config")
	}

	return kubeconfig, nil
}

// Exports cluster logs to a directory
func (m *Cluster) Logs(ctx context.Context) (*dagger.Directory, error) {
	dir := filepath.Join("tmp", m.Name, "logs")

	cmd := []string{"kind", "export", "logs", dir}

	if m.Name != "kind" {
		cmd = append(cmd, "--name", m.Name)
	}

	ctr, err := m.Kind.exec(ctx, m.Network, cmd...)
	if err != nil {
		return nil, err
	}

	return ctr.Directory(dir), nil
}

// Delete deletes the cluster if it exists.
func (m *Cluster) Delete(ctx context.Context) (string, error) {
	if !m.Exists {
		return fmt.Sprintf("cluster %s doesn't exist", m.Name), nil
	}

	cmd := []string{"kind", "delete", "cluster"}

	if m.Name != "kind" {
		cmd = append(cmd, "--name", m.Name)
	}

	_, err := m.Kind.exec(ctx, m.Network, cmd...)
	if err != nil {
		return "", err
	}

	m.Network = ""
	m.Exists = false

	return fmt.Sprintf("cluster %s deleted", m.Name), nil
}

// getContainerNetwork returns the network name of the given container.
func (k *Kind) getContainerNetwork(ctx context.Context, name string) (string, error) {
	return dag.Docker().
		Cli(dagger.DockerCliOpts{Socket: k.DockerSocket}).
		Run(ctx, []string{"container", "ls", "--filter", fmt.Sprintf("name=%s", name), "--format", "{{.Networks}}", "-n", "1"})
}

// helper functions

func (k *Kind) isClusterExist(ctx context.Context, name string) (bool, error) {
	cli := k.Container()

	clusters, err := cli.WithExec([]string{"kind", "get", "clusters"}).Stdout(ctx)
	if err != nil {
		return false, err
	}

	return lo.Contains(strings.Split(clusters, "\n"), name), nil
}

func (k *Kind) exec(ctx context.Context, kindNetwork string, args ...string) (*dagger.Container, error) {
	return k.Container().
		WithEnvVariable("KIND_EXPERIMENTAL_DOCKER_NETWORK", kindNetwork).
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithExec(args).
		Sync(ctx)
}

func (k *Kind) getClusterIPAddress(ctx context.Context, network, name string) (string, error) {
	return dag.Docker().
		Cli(dagger.DockerCliOpts{Socket: k.DockerSocket}).
		Run(ctx, []string{"inspect", fmt.Sprintf("%s-control-plane", name), "--format", fmt.Sprintf("{{.NetworkSettings.Networks.%s.IPAddress}}", network)})
}
