package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
)

// asJSON unmarshal the stdout of the container as JSON into the given value.
func (c *Container) asJSON(ctx context.Context, v interface{}) error {
	stdout, err := c.Stdout(ctx)
	if err != nil {
		return fmt.Errorf("%w: failed to get stdout", err)
	}

	err = json.Unmarshal([]byte(stdout), v)
	if err != nil {
		return fmt.Errorf("%w: failed to unmarshal stdout", err)
	}

	return nil
}

// unmarshalContentsToJSON unmarshal the contents of the file as JSON into the given value.
func (f *File) unmarshalContentsToJSON(ctx context.Context, v interface{}) error {
	stdout, err := f.Contents(ctx)
	if err != nil {
		return fmt.Errorf("%w: failed to get file contents", err)
	}

	err = json.Unmarshal([]byte(stdout), v)
	if err != nil {
		return fmt.Errorf("%w: failed to unmarshal file contents", err)
	}

	return nil
}

// unmarshalContentsToYAML unmarshal the contents of the file as YAML into the given value.
func (f *File) unmarshalContentsToYAML(ctx context.Context, v interface{}) error {
	stdout, err := f.Contents(ctx)
	if err != nil {
		return fmt.Errorf("%w: failed to get file contents", err)
	}

	err = yaml.Unmarshal([]byte(stdout), v)
	if err != nil {
		return fmt.Errorf("%w: failed to unmarshal file contents", err)
	}

	return nil
}

const GHCliVersion = "v2.24.0"

// withGithubCli installs the github cli in the container.
func (c *Container) withGithubCli() *Container {
	return c.WithFile("/usr/local/bin/gh", dag.Gh().GetGithubCli(GHCliVersion))
}

// gh returns a container with the github cli as the entrypoint.
func gh() *Container {
	return git().withGithubCli().WithEntrypoint([]string{"/usr/local/bin/gh"})
}

// git returns a container with the git image.
func git() *Container {
	return dag.Container().From("alpine/git:latest")
}

// base returns a container with the base image for gale.
func base() *Container {
	return dag.Container().From("ghcr.io/aweris/gale:v0.0.0-zenith")
}

// withGHX loads the necessary gale configuration into the container.
func (c *Container) withGHX(ctx context.Context) (*Container, error) {
	container := c

	ghx := base().File("/usr/local/bin/ghx")

	if _, err := ghx.Size(ctx); err != nil {
		return nil, fmt.Errorf("failed to get gale executor")
	}

	container = container.WithFile("/usr/local/bin/ghx", ghx)

	return container, nil
}
