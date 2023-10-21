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
	if opts.Repo != "" && opts.Branch == "" && opts.Commit == "" && opts.Tag == "" {
		return nil, fmt.Errorf("when repo is provided, one of branch, commit or tag must be provided")
	}

	source, err := g.getGithubRepositorySource(opts)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get repository source", err)
	}

	opts.Source = source

	info, err := g.getGithubRepository(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get repository information", err)
	}

	ref, err := g.getRepositoryRef(ctx, info, source, opts)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get repository ref", err)
	}

	return &Repo{Config: &RepoConfig{Info: info, Source: source, Ref: ref}}, nil
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
