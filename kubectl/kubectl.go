package main

import (
	"context"
	"fmt"
	"time"
)

// Kubectl represents the kubectl module for Dagger.
type Kubectl struct {
	Config *File
}

// WithConfig sets the kubectl config to use.
func (m *Kubectl) WithConfig(config *File) *Kubectl {
	m.Config = config
	return m
}

// WithRawConfig sets the kubectl config to use with the given contents.
func (m *Kubectl) WithRawConfig(contents string) *Kubectl {
	m.Config = dag.Directory().WithNewFile("config", contents).File("config")

	return m
}

// Exec executes the given kubectl command and returns the stdout.
func (m *Kubectl) Exec(ctx context.Context, args []string) (string, error) {
	if m.Config == nil {
		return "", fmt.Errorf("please provide a kubectl config")
	}

	return m.container().WithExec(args).Stdout(ctx)
}

// Container returns a container with the kubectl image and given config. The entrypoint is set to kubectl.
func (m *Kubectl) Container(_ context.Context) (*Container, error) {
	if m.Config == nil {
		return nil, fmt.Errorf("please provide a kubectl config")
	}

	return m.container(), nil
}

// container returns a container with the kubectl image.
func (m *Kubectl) container() *Container {
	return dag.Container().
		From("bitnami/kubectl:latest").
		WithUser("root").
		WithMountedFile("/.kube/config", m.Config).
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano))
}
