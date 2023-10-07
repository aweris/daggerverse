package main

import (
	"context"
	"fmt"
	"time"
)

// Helm represents the Helm module for Dagger.
type Helm struct{}

// Cli returns a Helm CLI with the given config.
func (m *Helm) Cli(config *File) *Cli {
	return &Cli{Config: config}
}

type Cli struct {
	Config *File
}

// Exec executes the given Helm command and returns the stdout.
func (m *Cli) Exec(ctx context.Context, args []string) (string, error) {
	if m.Config == nil {
		return "", fmt.Errorf("please provide a Helm config")
	}

	return m.container().WithExec(args).Stdout(ctx)
}

// Container returns a container with the Helm image and given config. The entrypoint is set to Helm.
func (m *Cli) Container(_ context.Context) (*Container, error) {
	if m.Config == nil {
		return nil, fmt.Errorf("please provide a Helm config")
	}

	return m.container(), nil
}

// container returns a container with the Helm image.
func (m *Cli) container() *Container {
	return dag.Container().
		From("alpine/helm:latest").
		WithUser("root").
		WithFile("/root/.kube/config", m.Config, ContainerWithFileOpts{Permissions: 0600}).
		WithMountedCache("/root/.cache/helm", dag.CacheVolume("helm-cache")).
		WithMountedCache("/root/.helm", dag.CacheVolume("helm-root")).
		WithMountedCache("/root/.config/helm", dag.CacheVolume("helm-config")).
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano))
}
