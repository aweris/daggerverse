package main

import (
	"context"
	"fmt"
	"path/filepath"
)

func (g *Gale) Runner() *Runner {
	return new(Runner)
}

type Runner struct {
	Base   *Container
	Config *Config
}

func (r *Runner) From(address string) *Runner {
	return &Runner{
		Base:   dag.Container().From(address),
		Config: r.Config,
	}
}

func (r *Runner) WithContainer(container *Container) *Runner {
	return &Runner{
		Base:   container,
		Config: r.Config,
	}
}

func (r *Runner) WithRepo(opts RepoOpts) *Runner {
	cfg := &Config{
		RepoOpts:   &opts,
		GithubOpts: r.Config.GithubOpts,
		EventOpts:  r.Config.EventOpts,
	}

	return &Runner{
		Base:   r.Base,
		Config: cfg,
	}
}

func (r *Runner) WithGithub(opts GithubOpts) *Runner {
	cfg := &Config{
		RepoOpts:   r.Config.RepoOpts,
		GithubOpts: &opts,
		EventOpts:  r.Config.EventOpts,
	}

	return &Runner{
		Base:   r.Base,
		Config: cfg,
	}
}

func (r *Runner) WithEvent(opts EventOpts) *Runner {
	cfg := &Config{
		RepoOpts:   r.Config.RepoOpts,
		GithubOpts: r.Config.GithubOpts,
		EventOpts:  &opts,
	}

	return &Runner{
		Base:   r.Base,
		Config: cfg,
	}
}

func (r *Runner) Container(ctx context.Context) (container *Container, err error) {
	container = r.Base

	if r.Config == nil {
		r.Config = getDefaultConfig()
	}

	// load github config

	container, err = container.loadGithubConfig(ctx, r.Config.GithubOpts)
	if err != nil {
		return nil, err
	}

	// load repo config
	repo, err := r.loadRepo(ctx)
	if err != nil {
		return nil, err
	}

	container, err = container.loadRepoConfig(ctx, repo)
	if err != nil {
		return nil, err
	}

	// load event config
	container, err = container.loadEventConfig(ctx, r.Config.EventOpts)
	if err != nil {
		return nil, err
	}

	// load gale config

	// bind action services

	return container, nil
}

// loadGithubConfig loads the github configuration into the container.
func (c *Container) loadGithubConfig(ctx context.Context, github *GithubOpts) (*Container, error) {
	container := c

	// load url config from github config
	container = container.WithEnvVariable("GITHUB_API_URL", github.APIURL).
		WithEnvVariable("GITHUB_GRAPHQL_URL", github.GraphqlURL).
		WithEnvVariable("GITHUB_SERVER_URL", github.ServerURL)

	// validate token if it's not empty and set it as a secret. If it's empty, try to load it from the environment and
	// validate it. If it's not found in the environment, fail.
	if github.Token == "" {
		token, err := container.EnvVariable(ctx, "GITHUB_TOKEN")
		if err != nil {
			return nil, fmt.Errorf("%w: failed validating github token", err)
		}

		// we can't proceed without a token. This is limitation of the gale.
		if token == "" {
			return nil, fmt.Errorf("missing github token. Please set the GITHUB_TOKEN environment variable or pass it as an option")
		}
	} else {
		container = container.WithSecretVariable("GITHUB_TOKEN", dag.SetSecret("GITHUB_TOKEN", github.Token))
	}

	return container, nil
}

// loadEventConfig loads the event configuration into the container. context arg and error return value are not used
// here but added to keep the signature of the function consistent with other load*Config functions.
func (c *Container) loadEventConfig(_ context.Context, event *EventOpts) (*Container, error) {
	container := c

	container = container.WithEnvVariable("GITHUB_EVENT_NAME", event.EventName)

	if event.EventFile != nil {
		path := filepath.Join("/home", "runner", "work", "_temp", "_github_workflow", "event.json")

		container = container.WithMountedFile(path, event.EventFile).WithEnvVariable("GITHUB_EVENT_PATH", path)
	}

	return container, nil
}

// loadRepo loads the repository information into the container. context arg and error return value are not used here
// but added to keep the signature of the function consistent with other load*Config functions.
func (c *Container) loadRepoConfig(_ context.Context, repo *Repo) (*Container, error) {
	workdir := filepath.Join("/home", "runner", "work", repo.Info.Name, repo.Info.Name)

	container := c.WithMountedDirectory(workdir, repo.Source).
		WithWorkdir(workdir).
		WithEnvVariable("GITHUB_WORKSPACE", workdir).
		WithEnvVariable("GH_REPO", repo.Info.NameWithOwner). // go-gh respects this variable while loading the repository.
		WithEnvVariable("GITHUB_REPOSITORY", repo.Info.NameWithOwner).
		WithEnvVariable("GITHUB_REPOSITORY_ID", repo.Info.ID).
		WithEnvVariable("GITHUB_REPOSITORY_OWNER", repo.Info.Owner.Login).
		WithEnvVariable("GITHUB_REPOSITORY_OWNER_ID", repo.Info.Owner.ID).
		WithEnvVariable("GITHUB_REPOSITORY_URL", repo.Info.URL)

	return container, nil
}
