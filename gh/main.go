package main

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/hashicorp/go-getter"
)

type Gh struct{}

func (m *Gh) Run(ctx context.Context, version, token string, args []string) (string, error) {
	file, err := m.GetGithubCli(ctx, version)
	if err != nil {
		return "", err
	}

	args = append([]string{"/usr/local/bin/gh"}, args...)

	return dag.Container().
		From("alpine/git:latest").
		WithFile("/usr/local/bin/gh", file).
		WithEnvVariable("GITHUB_TOKEN", token).
		WithExec(args, ContainerWithExecOpts{SkipEntrypoint: true}).
		Stdout(ctx)
}

func (m *Gh) GetGithubCli(ctx context.Context, version string) (*File, error) {
	var (
		goos       = runtime.GOOS
		goarch     = runtime.GOARCH
		versionNum = strings.TrimPrefix(version, "v")
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
