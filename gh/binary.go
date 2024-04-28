package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/google/go-github/v61/github"
	"github.com/hashicorp/go-getter"
	"github.com/samber/lo"
)

// GHBinary is the configuration for the Github CLI binary.
type GHBinary struct {
	// Version of the Github CLI
	Version string

	// Operating system of the Github CLI
	GOOS string

	// Architecture of the Github CLI
	GOARCH string
}

// WithVersion returns the GHBinary with the given version.
func (b GHBinary) WithVersion(version string) GHBinary {
	return GHBinary{
		Version: version,
		GOOS:    b.GOOS,
		GOARCH:  b.GOARCH,
	}
}

// WithArch returns the GHBinary with the given architecture.
func (b GHBinary) WithArch(goarch string) GHBinary {
	return GHBinary{
		Version: b.Version,
		GOOS:    b.GOOS,
		GOARCH:  goarch,
	}
}

// WithOS returns the GHBinary with the given operating system.
func (b GHBinary) WithOS(goos string) GHBinary {
	return GHBinary{
		Version: b.Version,
		GOOS:    goos,
		GOARCH:  b.GOARCH,
	}
}

// getLatestCliVersion returns the latest version of the Github CLI.
func (b GHBinary) getLatestCliVersion(ctx context.Context) (string, error) {
	client := github.NewClient(nil)

	release, _, err := client.Repositories.GetLatestRelease(ctx, "cli", "cli")
	if err != nil {
		return "", err
	}

	return *release.TagName, nil
}

// binary returns the Github CLI binary.
func (b GHBinary) binary(ctx context.Context) (*File, error) {
	if b.Version == "" {
		version, err := b.getLatestCliVersion(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest GitHub CLI version: %w", err)
		}

		b.Version = version
	}

	var (
		goos    = lo.Ternary(b.GOOS != "", b.GOOS, runtime.GOOS)
		goarch  = lo.Ternary(b.GOARCH != "", b.GOARCH, runtime.GOARCH)
		version = strings.TrimPrefix(b.Version, "v")
		suffix  = lo.Ternary(goos == "linux", "tar.gz", "zip")
	)

	if goos != "linux" && goos != "darwin" {
		return nil, fmt.Errorf("unsupported operating system: %s", goos)
	}

	// github releases use "macOS" instead of "darwin"
	if goos == "darwin" {
		goos = "macOS"
	}

	src := fmt.Sprintf("https://github.com/cli/cli/releases/download/v%s/gh_%s_%s_%s.%s", version, version, goos, goarch, suffix)
	dst := fmt.Sprintf("gh_%s_%s_%s", version, goos, goarch)

	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	client := getter.Client{
		Ctx:  ctx,
		Src:  src,
		Dst:  pwd,
		Mode: getter.ClientModeDir,
	}

	err = client.Get()
	if err != nil {
		return nil, err
	}

	return dag.CurrentModule().WorkdirFile(path.Join(dst, "bin/gh")), nil
}
