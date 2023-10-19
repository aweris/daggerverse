package main

import (
	"context"
	"fmt"
)

// Repo represents a Github repository and its source.
type Repo struct {
	Config *RepoConfig
}

// Repo returns a new Repo with given repository options.
func (g *Gale) Repo(ctx context.Context, opts RepoOpts) (*Repo, error) {
	if g.Config.Token == nil {
		return nil, fmt.Errorf("github token must be provided")
	}

	if opts.Repo == "" && opts.Source == nil {
		return nil, fmt.Errorf("repository name or source must be provided")
	}

	info, err := g.getGithubRepository(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get repository information", err)
	}

	source, err := g.getGithubRepositorySource(info, opts)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get repository source", err)
	}

	ref, err := g.getRepositoryRef(ctx, info, source, opts)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get repository ref", err)
	}

	return &Repo{Config: &RepoConfig{AuthConfig: g.Config, Info: info, Source: source, Ref: ref}}, nil
}

// Workflows returns a workflows of the repository at given path.
func (r *Repo) Workflows(opts WorkflowOpts) *Workflows {
	return &Workflows{
		Config: &WorkflowsConfig{
			RepoConfig:   r.Config,
			WorkflowsDir: opts.WorkflowsDir,
		},
	}
}
