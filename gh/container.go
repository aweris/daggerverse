package main

import (
	"github.com/samber/lo"
)

type GHContainer struct {
	// Base container for the Github CLI
	Base *Container

	// Github token
	Token *Secret

	// Github Repository
	Repo string
}

// WithRepo returns the GHContainer with the given repository.
func (c GHContainer) WithRepo(repo string) GHContainer {
	return GHContainer{
		Base:  c.Base,
		Token: c.Token,
		Repo:  repo,
	}
}

// WithToken returns the GHContainer with the given token.
func (c GHContainer) WithToken(token *Secret) GHContainer {
	return GHContainer{
		Base:  c.Base,
		Token: token,
		Repo:  c.Repo,
	}
}

// container returns the container for the Github CLI with the given binary.
func (c GHContainer) container(binary *File) *Container {
	return lo.Ternary(c.Base != nil, c.Base, dag.Container().From("alpine/git:latest")).
		WithFile("/usr/local/bin/gh", binary).
		WithEntrypoint([]string{"/usr/local/bin/gh"}).
		WithEnvVariable("GH_PROMPT_DISABLED", "true").
		WithEnvVariable("GH_NO_UPDATE_NOTIFIER", "true").
		With(func(ctr *Container) *Container {
			if c.Token != nil {
				ctr = ctr.WithSecretVariable("GH_TOKEN", c.Token)
			}

			if c.Repo != "" {
				ctr = ctr.WithEnvVariable("GH_REPO", c.Repo)
			}

			return ctr
		})
}
