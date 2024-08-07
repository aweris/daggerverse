// A Dagger Module for integrating with the KinD
package main

import (
	"context"

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

// Returns a cluster object that can be used to interact with the kind cluster
func (k *Kind) Cluster(
	ctx context.Context,

	// Name of the cluster
	//
	// +optional
	// +default="kind"
	name string,
) (*Cluster, error) {
	// Get the network name for the engine containers to ensure the cluster is created on the same network. It's
	// important to use the same network to be able to access the cluster from other containers using the IP address of
	// the cluster.
	network, err := getContainerNetwork(ctx, k.DockerSocket, "^dagger-engine-*")
	if err != nil {
		return nil, err
	}

	return &Cluster{Name: name, Network: network, Kind: k}, nil
}
