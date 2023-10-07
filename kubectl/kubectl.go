package main

import (
	"context"
	"fmt"
	"time"
)

// Kubectl represents the kubectl module for Dagger.
type Kubectl struct{}

// Cli returns a kubectl cli with the given config.
func (m *Kubectl) Cli(config *File) *Cli {
	return &Cli{Config: config}
}

type Cli struct {
	Config *File
}

// Exec executes the given kubectl command and returns the stdout.
func (m *Cli) Exec(ctx context.Context, args []string) (string, error) {
	if m.Config == nil {
		return "", fmt.Errorf("please provide a kubectl config")
	}

	return m.container().WithExec(args).Stdout(ctx)
}

// Container returns a container with the kubectl image and given config. The entrypoint is set to kubectl.
func (m *Cli) Container(_ context.Context) (*Container, error) {
	if m.Config == nil {
		return nil, fmt.Errorf("please provide a kubectl config")
	}

	return m.container(), nil
}

// container returns a container with the kubectl image.
func (m *Cli) container() *Container {
	return dag.Container().
		From("bitnami/kubectl:latest").
		WithUser("root").
		WithMountedFile("/.kube/config", m.Config).
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano))
}
