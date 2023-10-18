package main

import (
	"context"
	"fmt"
)

// getGithubRepository returns a new GithubRepository with given repository name and options.
func (g *Gale) getGithubRepository(ctx context.Context, opts RepoOpts) (*GithubRepository, error) {
	args := []string{"repo", "view", opts.Repo, "--json", "id,name,owner,nameWithOwner,url,defaultBranchRef"}

	container := gh().WithSecretVariable("GITHUB_TOKEN", g.Config.Token)

	// if the repository is not set, mount the current directory as the repository
	if opts.Repo == "" {
		container = container.WithMountedDirectory("/src", opts.Source).WithWorkdir("/src")
	}

	var info *GithubRepository

	err := container.WithExec(args).asJSON(ctx, &info)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get repository information", err)
	}

	return info, nil
}

// getGithubRepositorySource returns a new Directory with given repository information and options. If all options are
// empty, the current directory is returned. If repo name is provided but no other options are provided, the default
// branch is used.
func (g *Gale) getGithubRepositorySource(info *GithubRepository, opts RepoOpts) (*Directory, error) {
	source := opts.Source

	// if all options are empty, use the current directory
	if opts.Repo == "" && opts.Branch == "" && opts.Commit == "" && opts.Tag == "" {
		return dag.Host().Directory("."), nil
	}

	// if all options are empty, find default branch and use that
	if opts.Tag == "" && opts.Branch == "" && opts.Commit == "" {
		opts.Branch = info.DefaultBranchRef.Name
	}

	switch {
	case opts.Tag != "":
		source = dag.Git(info.URL, GitOpts{KeepGitDir: true}).Tag(opts.Tag).Tree()
	case opts.Branch != "":
		source = dag.Git(info.URL, GitOpts{KeepGitDir: true}).Branch(opts.Branch).Tree()
	case opts.Commit != "":
		source = dag.Git(info.URL, GitOpts{KeepGitDir: true}).Commit(opts.Commit).Tree()
	default:
		return nil, fmt.Errorf("couldn't find a repository to load") // this should never happen, added for defensive programming
	}

	return source, nil
}
