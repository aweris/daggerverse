package main

import (
	"context"
	"time"
)

// Gale is a Dagger module for running Github Actions workflows.
type Gale struct{}

// Repo returns a Repo with the given name.
func (gale *Gale) Repo(_ context.Context, name string) (*Repo, error) {
	return &Repo{RepoName: name}, nil
}

// Repo represents a Github repository to use.
type Repo struct {
	RepoName     string
	WorkflowsDir string
	Branch       string
	Tag          string
}

// WithWorkflowsDir sets the workflows directory. If not set, the default is ".github/workflows".
func (repo *Repo) WithWorkflowsDir(_ context.Context, dir string) (*Repo, error) {
	repo.WorkflowsDir = dir
	return repo, nil
}

// WithBranch sets the branch to use for the repository. If not set, the default is the default branch of the repository.
func (repo *Repo) WithBranch(_ context.Context, branch string) (*Repo, error) {
	repo.Branch = branch
	return repo, nil
}

// WithTag sets the tag to use for the repository. If not set, the default is the default branch of the repository.
func (repo *Repo) WithTag(_ context.Context, tag string) (*Repo, error) {
	repo.Tag = tag
	return repo, nil
}

// List lists the workflows in the repository.
func (repo *Repo) List(ctx context.Context, token string) (string, error) {
	exec := []string{"list", "--repo", repo.RepoName}

	if repo.WorkflowsDir != "" {
		exec = append(exec, "--workflows-dir", repo.WorkflowsDir)
	} else {
		exec = append(exec, "--workflows-dir", ".github/workflows") //  FIXME: in gale
	}

	if repo.Branch != "" {
		exec = append(exec, "--branch", repo.Branch)
	}

	if repo.Tag != "" {
		exec = append(exec, "--tag", repo.Tag)
	}

	return run(ctx, token, exec)
}

// Workflow returns a Workflow with the given name.
func (repo *Repo) Workflow(_ context.Context, name string) (*Workflow, error) {
	return &Workflow{
		RepoName:     repo.RepoName,
		WorkflowsDir: repo.WorkflowsDir,
		Branch:       repo.Branch,
		Tag:          repo.Tag,
		WorkflowName: name,
	}, nil
}

// Workflow represents a Github workflow to run.
type Workflow struct {
	RepoName     string
	WorkflowsDir string
	Branch       string
	Tag          string
	WorkflowName string
	Job          string
	Debug        bool
}

// WithJob sets the job to run. If not set, the default is running all jobs in the workflow.
func (workflow *Workflow) WithJob(_ context.Context, name string) (*Workflow, error) {
	workflow.Job = name
	return workflow, nil
}

// WithDebug sets the debug mode. If not set, the default is false.
func (workflow *Workflow) WithDebug(_ context.Context) (*Workflow, error) {
	workflow.Debug = true
	return workflow, nil
}

// Run runs the workflow.
func (workflow *Workflow) Run(ctx context.Context, token string) (string, error) {
	exec := []string{"run", "--repo", workflow.RepoName}

	if workflow.WorkflowsDir != "" {
		exec = append(exec, "--workflows-dir", workflow.WorkflowsDir)
	} else {
		exec = append(exec, "--workflows-dir", ".github/workflows") //  FIXME: in gale
	}

	if workflow.Branch != "" {
		exec = append(exec, "--branch", workflow.Branch)
	}

	if workflow.Tag != "" {
		exec = append(exec, "--tag", workflow.Tag)
	}

	if workflow.Debug {
		exec = append(exec, "--debug")
	}

	exec = append(exec, workflow.WorkflowName)

	if workflow.Job != "" {
		exec = append(exec, "--job", workflow.Job)
	}

	return run(ctx, token, exec)
}

// run runs the given exec command on the gale container with the given token as the GITHUB_TOKEN environment variable
// and returns the stdout of the command.
func run(ctx context.Context, token string, exec []string) (string, error) {
	return dag.Container().
		From("ghcr.io/aweris/gale:v0.0.0-zenith").
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
		WithSecretVariable("GITHUB_TOKEN", dag.SetSecret("GITHUB_TOKEN", token)).
		WithExec(exec, ContainerWithExecOpts{ExperimentalPrivilegedNesting: true}).
		Stdout(ctx)
}
