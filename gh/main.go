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

	// Configuration for the Github CLI container
	// +private
	GHContainer GHContainer
}

func New(
	// GitHub CLI version. (default: latest version)
	// +optional
	version string,

	// GitHub token.
	// +optional
	token *Secret,

	// GitHub repository (e.g. "owner/repo").
	// +optional
	repo string,

	// Base container for the Github CLI
	// +optional
	base *Container,
) *Gh {
	return &Gh{
		Binary: GHBinary{
			Version: version,
		},
		GHContainer: GHContainer{
			Base:  base,
			Token: token,
			Repo:  repo,
		},
	}
}

func (m *Gh) Container(
	ctx context.Context,

	// GitHub CLI version. (default: latest version)
	// +optional
	version string,

	// GitHub token.
	// +optional
	token *Secret,

	// GitHub repository (e.g. "owner/repo").
	// +optional
	repo string,
) (*Container, error) {
	file, err := lo.Ternary(version != "", m.Binary.WithVersion(version), m.Binary).binary(ctx)
	if err != nil {
		return nil, err
	}

	// get the github container configuration
	gc := m.GHContainer

	// update the container with the given token and repository if provided
	gc = lo.Ternary(token != nil, gc.WithToken(token), gc)
	gc = lo.Ternary(repo != "", gc.WithRepo(repo), gc)

	// get the container object with the given binary
	ctr := gc.container(file)

	return ctr, nil
}

// Run a GitHub CLI command (accepts a single command string without "gh").
func (m *Gh) Run(
	ctx context.Context,

	// Command to run.
	cmd string,

	// GitHub CLI version. (default: latest version)
	// +optional
	version string,

	// GitHub token.
	// +optional
	token *Secret,

	// GitHub repository (e.g. "owner/repo").
	// +optional
	repo string,

	// disable cache
	// +optional
	// +default=false
	disableCache bool,
) (*Container, error) {
	ctr, err := m.Container(ctx, version, token, repo)
	if err != nil {
		return nil, err
	}

	// disable cache if requested
	ctr = lo.Ternary(disableCache, ctr.WithEnvVariable("CACHE_BUSTER", time.Now().String()), ctr)

	// run the command and return the container
	return ctr.WithExec([]string{"sh", "-c", strings.Join([]string{"/usr/local/bin/gh", cmd}, " ")}, ContainerWithExecOpts{SkipEntrypoint: true}), nil
}

// Get the GitHub CLI binary.
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
