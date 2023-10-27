package main

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/hashicorp/go-getter"
)

const (
	DefaultCLIVersion = "v2.37.0"
)

// Gh is Github CLI module for Dagger
type Gh struct{}

func (m *Gh) Run(
	ctx context.Context,
	// version of the Github CLI (default: v2.37.0)
	version Optional[string],
	// Github token
	token *Secret,
	// command to run
	cmd string,
) (string, error) {
	file, err := m.Get(ctx, version)
	if err != nil {
		return "", err
	}

	return dag.Container().
		From("alpine/git:latest").
		WithFile("/usr/local/bin/gh", file).
		WithSecretVariable("GITHUB_TOKEN", token).
		WithExec([]string{"sh", "-c", strings.Join([]string{"/usr/local/bin/gh", cmd}, " ")}, ContainerWithExecOpts{SkipEntrypoint: true}).
		Stdout(ctx)
}

// Get returns the Github CLI binary
func (m *Gh) Get(
	ctx context.Context,
	// version of the Github CLI (default: v2.37.0)
	version Optional[string],
) (*File, error) {
	var (
		goos       = runtime.GOOS
		goarch     = runtime.GOARCH
		versionNum = strings.TrimPrefix(version.GetOr(DefaultCLIVersion), "v")
	)

	src := fmt.Sprintf("https://github.com/cli/cli/releases/download/v%s/gh_%s_%s_%s.tar.gz", versionNum, versionNum, goos, goarch)
	dst := fmt.Sprintf("/tmp/gh_%s_%s_%s", versionNum, goos, goarch)

	client := getter.Client{
		Ctx:  ctx,
		Src:  src,
		Dst:  "/tmp",
		Mode: getter.ClientModeDir,
	}

	err := client.Get()
	if err != nil {
		return nil, err
	}

	dir := dag.Host().Directory(dst)

	return dir.File("bin/gh"), nil
}
