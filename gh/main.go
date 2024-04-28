package main

import (
	"context"
	"strings"
	"time"

	"github.com/samber/lo"
)

// Gh is Github CLI module for Dagger
type Gh struct {
	// Configuration for the Github CLI binary
	// +private
	Binary GHBinary
}

func New(
	// version of the Github CLI
	// +optional
	version string,
) *Gh {
	return &Gh{
		Binary: GHBinary{
			Version: version,
		},
	}
}

func (m *Gh) Run(
	ctx context.Context,
	// Github token
	token *Secret,
	// command to run
	cmd string,
	// version of the Github CLI
	// +optional
	// +default="2.37.0"
	version string,
	// disable cache
	// +optional
	// +default=false
	disableCache bool,
) (string, error) {
	file, err := lo.Ternary(version != "", m.Binary.WithVersion(version), m.Binary).binary(ctx)
	if err != nil {
		return "", err
	}

	ctr := dag.Container().
		From("alpine/git:latest").
		WithFile("/usr/local/bin/gh", file).
		WithSecretVariable("GITHUB_TOKEN", token)

	if disableCache {
		ctr = ctr.WithEnvVariable("CACHEBUSTER", time.Now().String())
	}

	return ctr.WithExec([]string{"sh", "-c", strings.Join([]string{"/usr/local/bin/gh", cmd}, " ")}, ContainerWithExecOpts{SkipEntrypoint: true}).
		Stdout(ctx)
}

// Get returns the Github CLI binary
func (m *Gh) Get(
	ctx context.Context,
	// operating system of the binary
	// +optional
	goos string,
	// architecture of the binary
	// +optional
	goarch string,
	// version of the Github CLI
	// +optional
	version string,
) (*File, error) {
	return lo.Ternary(version != "", m.Binary.WithVersion(version), m.Binary).
		WithOS(goos).
		WithArch(goarch).
		binary(ctx)
}
