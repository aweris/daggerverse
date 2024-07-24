package main

import (
	"context"
	"fmt"
	"time"

	"dagger/kind/internal/dagger"
)

// Helper Functions

func getContainerNetwork(ctx context.Context, socket *dagger.Socket, name string) (string, error) {
	return dag.Docker().
		Cli(dagger.DockerCliOpts{Socket: socket}).
		Run(ctx, []string{"container", "ls", "--filter", fmt.Sprintf("name=%s", name), "--format", "{{.Networks}}", "-n", "1"})
}

func exec(ctx context.Context, ctr *dagger.Container, kindNetwork string, args ...string) (*dagger.Container, error) {
	return ctr.
		WithEnvVariable("KIND_EXPERIMENTAL_DOCKER_NETWORK", kindNetwork).
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithExec(args).
		Sync(ctx)
}

func getClusterIPAddress(ctx context.Context, socket *dagger.Socket, network, name string) (string, error) {
	return dag.Docker().
		Cli(dagger.DockerCliOpts{Socket: socket}).
		Run(ctx, []string{"inspect", fmt.Sprintf("%s-control-plane", name), "--format", fmt.Sprintf("{{.NetworkSettings.Networks.%s.IPAddress}}", network)})
}
