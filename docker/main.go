// A Dagger Module for integrating with the Docker Engine
package main

import (
	"context"
	"docker/internal/dagger"
	"fmt"
	"strings"
)

// A Dagger module to integrate with Docker
type Docker struct{}

// Spawn an ephemeral Docker Engine in a container
func (e *Docker) Engine(
	// Docker Engine version
	//
	// +optional
	// +default="26.1"
	version string,

	// Persist the state of the engine in a cache volume
	//
	// +optional
	// +default=true
	persist bool,

	// Namespace for persisting the engine state. Use in combination with `persist`
	//
	// +optional
	namespace string,
) *dagger.Service {
	ctr := dag.Container().From(fmt.Sprintf("index.docker.io/docker:%s-dind", version))

	// disable the entrypoint
	ctr = ctr.WithoutEntrypoint()

	// expose the Docker Engine API
	ctr = ctr.WithExposedPort(2375)

	// persist the engine state
	if persist {
		var (
			name   = strings.TrimSuffix("docker-engine-state-"+version+"-"+namespace, "-")
			volume = dag.CacheVolume(name)
			opts   = dagger.ContainerWithMountedCacheOpts{Sharing: dagger.Locked}
		)

		ctr = ctr.WithMountedCache("/var/lib/docker", volume, opts)
	}

	return ctr.
		WithExec(
			[]string{
				"dockerd",
				"--host=tcp://0.0.0.0:2375",
				"--host=unix:///var/run/docker.sock",
				"--tls=false",
			},
			dagger.ContainerWithExecOpts{InsecureRootCapabilities: true},
		).
		AsService()
}

// A Docker CLI ready to query this engine.
// Entrypoint is set to `docker`
func (d *Docker) CLI(
	// Version of the Docker CLI to run.
	//
	// +optional
	// +default="26.1"
	version string,

	// Specify the Docker Engine to connect to. By default, run an ephemeral engine.
	//
	// +optional
	engine *dagger.Service,

	// Unix socket to connect to the external Docker Engine.Please carefully use this option it can expose your host to the container.
	//
	// +optional
	socket *dagger.Socket,
) *CLI {

	if socket == nil && engine == nil {
		engine = d.Engine(version, true, "default")
	}

	return &CLI{
		Engine: engine,
		Socket: socket,
	}
}

// A Docker client
type CLI struct {
	// +private
	Engine *dagger.Service

	// +private
	Socket *dagger.Socket
}

// Package the Docker CLI into a container, wired to an engine
func (c *CLI) Container() *dagger.Container {
	ctr := dag.Container().From("index.docker.io/docker:cli")

	// disable the entrypoint
	ctr = ctr.WithoutEntrypoint()

	// wire the engine to the container
	switch {
	case c.Socket != nil:
		ctr = ctr.WithUnixSocket("/var/run/docker.sock", c.Socket)
		ctr = ctr.WithEnvVariable("DOCKER_HOST", "unix:///var/run/docker.sock")
	case c.Engine != nil:
		ctr = ctr.WithServiceBinding("dockerd", c.Engine)
		ctr = ctr.WithEnvVariable("DOCKER_HOST", "tcp://dockerd:2375")
	}

	return ctr
}

// Run a container with the docker CLI
func (c *CLI) Run(
	// Context for the command
	ctx context.Context,

	// Arguments to pass to the docker CLI command to run
	args []string,
) (string, error) {
	out, err := c.Container().WithExec(append([]string{"docker"}, args...)).Stdout(ctx)

	return strings.TrimSpace(out), err
}
