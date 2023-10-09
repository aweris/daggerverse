package main

import (
    "context"
    "time"
)

// Gale is a Dagger module for running Github Actions workflows.
type Gale struct {
    debug bool
    token *Secret
}

// List lists the workflows in the repository.
func (g *Gale) List(ctx context.Context, token string, opts RepoOpts) (string, error) {
    // set gale options first
    g.debug = false
    g.token = dag.SetSecret("GITHUB_TOKEN", token)

    // create the exec command for listing workflows
    exec := []string{"list"}

    // append repo options to the exec command
    exec = append(exec, opts.toExecArgs()...)

    container, err := g.run(ctx, exec)
    if err != nil {
        return "", err
    }

    return container.Stdout(ctx)
}

// Run runs the workflow.
func (g *Gale) Run(ctx context.Context, workflow, token string, repoOpts RepoOpts, runOpts RunOpts, runnerOpts RunnerOpts) (string, error) {
    // set gale options first
    g.debug = runnerOpts.Debug
    g.token = dag.SetSecret("GITHUB_TOKEN", token)

    // create the exec command for running the workflow
    exec := []string{"run"}

    // append repo options to the exec command
    exec = append(exec, repoOpts.toExecArgs()...)

    // append workflow to the exec command
    exec = append(exec, workflow)

    // append run options to the exec command
    exec = append(exec, runOpts.toExecArgs()...)

    container, err := g.run(ctx, exec)
    if err != nil {
        return "", err
    }

    return container.Stdout(ctx)
}

// container returns the gale container.
func (g *Gale) container() *Container {
    return dag.Container().
        From("ghcr.io/aweris/gale:v0.0.0-zenith").
        withRunnerDebug(g.debug).
        withGithubToken(g.token).
        WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano))
}

// withRunnerDebug sets the RUNNER_DEBUG environment variable 1 if debug is true. This is used to enable debug mode
// in for the Github Actions workflows.
func (c *Container) withRunnerDebug(debug bool) *Container {
    if debug {
        return c.WithEnvVariable("RUNNER_DEBUG", "1")
    }

    return c
}

// withGithubToken sets the GITHUB_TOKEN secret if token is not empty. This is used to authenticate the Github Actions
// workflows.
func (c *Container) withGithubToken(token *Secret) *Container {
    if token != nil {
        return c.WithSecretVariable("GITHUB_TOKEN", token)
    }

    return c
}

// run runs the gale container with the given exec command.
func (g *Gale) run(ctx context.Context, exec []string) (*Container, error) {
    container := g.container()

    if endpoint, service, err := g.serviceArtifacts(ctx); err == nil {
        container = container.WithServiceBinding("artifacts", service).WithEnvVariable("ACTIONS_RUNTIME_URL", endpoint)
    } else {
        return nil, err
    }

    if endpoint, service, err := g.serviceArtifactCache(ctx); err == nil {
        container = container.WithServiceBinding("artifactcache", service).WithEnvVariable("ACTIONS_CACHE_URL", endpoint)
    } else {
        return nil, err
    }

    container = container.WithEnvVariable("ACTIONS_RUNTIME_TOKEN", "token")

    return container.WithExec(exec, ContainerWithExecOpts{ExperimentalPrivilegedNesting: true}), nil
}
