package main

import (
	"context"
	"fmt"
	"runtime"
	"strings"
)

// Gh is Github CLI module for Dagger
type Gh struct{}

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
	disableCache bool
) (string, error) {
	file := m.Get(ctx, version)

	ctr := dag.Container().
		From("alpine/git:latest").
		WithFile("/usr/local/bin/gh", file).
		WithSecretVariable("GITHUB_TOKEN", token)
	
	if (disableCache) {
		ctr = ctr.WithEnvVariable("CACHEBUSTER", time.Now().String())
	}

	return ctr.WithExec([]string{"sh", "-c", strings.Join([]string{"/usr/local/bin/gh", cmd}, " ")}, ContainerWithExecOpts{SkipEntrypoint: true}).
		Stdout(ctx)
}

// Get returns the Github CLI binary
func (m *Gh) Get(
	ctx context.Context,
	// version of the Github CLI
	// +optional
	// +default="2.37.0"
	version string,
) *File {
	var (
		goos       = runtime.GOOS
		goarch     = runtime.GOARCH
		versionNum = version
	)

	src := fmt.Sprintf("https://github.com/cli/cli/releases/download/v%s/gh_%s_%s_%s.tar.gz", versionNum, versionNum, goos, goarch)
	dst := fmt.Sprintf("gh_%s_%s_%s", versionNum, goos, goarch)

	return dag.Container().From("alpine").
		WithMountedFile("/tmp/gh.tar.gz", dag.HTTP(src)).
		WithWorkdir("/tmp").
		WithExec([]string{"tar", "xvf", "gh.tar.gz"}).
		File(fmt.Sprintf("%s/bin/gh", dst))
}
