package main

import (
	"context"
	"encoding/json"
	"fmt"
)

type Repo struct {
	Info   *GithubRepository
	Source *Directory
}

// GithubRepository represents a GitHub repository
type GithubRepository struct {
	ID               string                    `json:"id"`
	Name             string                    `json:"name"`
	NameWithOwner    string                    `json:"name_with_owner"`
	URL              string                    `json:"url" `
	Owner            GithubRepositoryOwner     `json:"owner"`
	DefaultBranchRef GithubRepositoryBranchRef `json:"default_branch_ref"`
}

// GithubRepositoryOwner represents a GitHub repository owner
type GithubRepositoryOwner struct {
	ID    string `json:"id" env:"GALE_REPO_OWNER_ID" container_env:"true"`
	Login string `json:"login" env:"GALE_REPO_OWNER_LOGIN" container_env:"true"`
}

// GithubRepositoryBranchRef represents a GitHub repository branch ref
type GithubRepositoryBranchRef struct {
	Name string `json:"name" env:"GALE_REPO_BRANCH_NAME" container_env:"true"`
}

func (r *Runner) loadRepo(ctx context.Context) (*Repo, error) {
	var source *Directory

	args := []string{"repo", "view", r.Config.Repo, "--json", "id,name,owner,nameWithOwner,url,defaultBranchRef"}

	container := dag.Container().
		From("alpine/git:latest").
		WithFile("/usr/local/bin/gh", dag.Gh().GetGithubCli(GHCliVersion)).
		WithEntrypoint([]string{"/usr/local/bin/gh"}).
		WithSecretVariable("GITHUB_TOKEN", dag.SetSecret("GITHUB_TOKEN", r.Config.Token))

	// if the repository is not set, mount the current directory as the repository
	if r.Config.Repo == "" {
		source = dag.Host().Directory(".")

		container = container.WithMountedDirectory("/src", source).WithWorkdir("/src")
	}

	stdout, err := container.WithExec(args).Stdout(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get repository information", err)
	}

	var info *GithubRepository

	err = json.Unmarshal([]byte(stdout), &info)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse repository information", err)
	}

	// if all options are empty, find default branch and use that
	if r.Config.Tag == "" && r.Config.Branch == "" && r.Config.Commit == "" {
		r.Config.Branch = info.DefaultBranchRef.Name
	}

	switch {
	case r.Config.Tag != "":
		source = dag.Git(info.URL, GitOpts{KeepGitDir: true}).Tag(r.Config.Tag).Tree()
	case r.Config.Branch != "":
		source = dag.Git(info.URL, GitOpts{KeepGitDir: true}).Branch(r.Config.Branch).Tree()
	case r.Config.Commit != "":
		source = dag.Git(info.URL, GitOpts{KeepGitDir: true}).Commit(r.Config.Commit).Tree()
	case source != nil:
		// do nothing it's already set
	default:
		return nil, fmt.Errorf("couldn't find a repository to load")
	}

	return &Repo{Info: info, Source: source}, nil
}
