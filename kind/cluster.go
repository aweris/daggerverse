package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/samber/lo"

	"dagger/kind/internal/dagger"
)

// Represents a kind cluster
type Cluster struct {
	// Name of the cluster
	Name string

	// Network name of the cluster. This should be the same as the network name of the dagger-engine containers
	Network string

	// +private
	Kind *Kind
}

// Check if the cluster exists or not
func (c *Cluster) Exist(ctx context.Context) (bool, error) {
	ctr, err := exec(ctx, c.Container(), c.Network, []string{"kind", "get", "clusters"}...)
	if err != nil {
		return false, err
	}

	clusters, err := ctr.Stdout(ctx)
	if err != nil {
		return false, err
	}

	return lo.Contains(strings.Split(strings.TrimSpace(clusters), "\n"), c.Name), nil
}

// Create creates the cluster if it doesn't already exist.
func (c *Cluster) Create(ctx context.Context) (string, error) {
	exist, err := c.Exist(ctx)
	if err != nil {
		return "", err
	}

	if exist {
		currentNetwork, err := getContainerNetwork(ctx, c.Kind.DockerSocket, fmt.Sprintf("^%s-control-plane-*", c.Name))
		if err != nil {
			return "", err
		}

		// If the cluster is not on the same network as the engine containers, we'll return an error to avoid using the
		// cluster. This is important to be able to access the cluster from other containers using the IP address of
		// the cluster. If the cluster is not on the same network, the IP address of the cluster won't be accessible
		if currentNetwork != c.Network {
			return "", fmt.Errorf("cluster %s is not connected to the engine", c.Name)
		}

		return fmt.Sprintf("cluster %s already exists", c.Name), nil
	}

	cmd := []string{"kind", "create", "cluster"}

	_, err = exec(ctx, c.Container(), c.Network, cmd...)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("cluster %s created", c.Name), nil
}

// Delete deletes the cluster if it exists.
func (c *Cluster) Delete(ctx context.Context) (string, error) {
	exist, err := c.Exist(ctx)
	if err != nil {
		return "", err
	}

	if !exist {
		return fmt.Sprintf("cluster %s doesn't exist", c.Name), nil
	}

	cmd := []string{"kind", "delete", "cluster"}

	_, err = exec(ctx, c.Container(), c.Network, cmd...)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("cluster %s deleted", c.Name), nil
}

// Exports cluster logs to a directory
func (c *Cluster) Logs(ctx context.Context) (*dagger.Directory, error) {
	dir := filepath.Join("tmp", c.Name, "logs")

	cmd := []string{"kind", "export", "logs", dir}

	ctr, err := exec(ctx, c.Container(), c.Network, cmd...)
	if err != nil {
		return nil, err
	}

	return ctr.Directory(dir), nil
}

// Exports cluster kubeconfig
func (c *Cluster) Kubeconfig(
	ctx context.Context,

	// If true, the internal address is used in the kubeconfig. This is useful for running kubectl commands from within other containers.
	//
	// +optional
	// +default=false
	internal bool,
) (*dagger.File, error) {
	cmd := []string{"kind", "export", "kubeconfig"}

	if internal {
		cmd = append(cmd, "--internal")
	}

	ctr, err := exec(ctx, c.Container(), c.Network, cmd...)
	if err != nil {
		return nil, err
	}

	kubeconfig := ctr.Directory("/root/.kube").File("config")

	if internal {
		ip, err := getClusterIPAddress(ctx, c.Kind.DockerSocket, c.Network, c.Name)
		if err != nil {
			return nil, err
		}

		contents, err := kubeconfig.Contents(ctx)
		if err != nil {
			return nil, err
		}

		contents = strings.ReplaceAll(contents, fmt.Sprintf("https://%s-control-plane:6443", c.Name), fmt.Sprintf("https://%s:6443", ip))

		kubeconfig = dag.Directory().WithNewFile("config", contents).File("config")
	}

	return kubeconfig, nil
}

// Container that contains the kind and k9s binaries with the cluster name and network set as environment variables
func (c *Cluster) Container() *dagger.Container {
	return c.Kind.
		Container().
		WithEnvVariable("KIND_CLUSTER_NAME", c.Name).
		WithEnvVariable("KIND_EXPERIMENTAL_DOCKER_NETWORK", c.Network)
}
