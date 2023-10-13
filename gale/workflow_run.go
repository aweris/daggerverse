package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
)

type WorkflowRun struct {
	Config           *WorkflowRunConfig
	InternalServices *InternalServices
}

func (w *Workflows) Run(opts WorkflowRunOpts) *WorkflowRun {
	return &WorkflowRun{
		Config: &WorkflowRunConfig{
			WorkflowsConfig: w.Config,
			WorkflowRunOpts: &opts,
		},
		InternalServices: NewInternalServices(),
	}
}

// Result represents the result of a workflow run.
type Result struct {
	Ran         bool      `json:"ran"`         // Ran indicates if the execution ran
	Conclusion  string    `json:"conclusion"`  // Conclusion of the execution
	StartedAt   time.Time `json:"startedAt"`   // StartedAt time of the execution
	CompletedAt time.Time `json:"completedAt"` // CompletedAt time of the execution
}

// Sync forces to evaluate the workflow run and returns the container.
func (wr *WorkflowRun) Sync(ctx context.Context) (*Container, error) {
	container, err := wr.run(ctx)
	if err != nil {
		return nil, err
	}

	return container.Sync(ctx)
}

// Directory returns the directory of the workflow run information.
func (wr *WorkflowRun) Directory(ctx context.Context, opts WorkflowRunExportOpts) (*Directory, error) {
	container, err := wr.run(ctx)
	if err != nil {
		return nil, err
	}

	dir := dag.Directory().WithDirectory("run", container.Directory("/home/runner/_temp/ghx"))

	if opts.IncludeSource {
		dir = dir.WithDirectory("source", container.Directory(fmt.Sprintf("/home/runner/work/%s/%s", wr.Config.Info.Name, wr.Config.Info.Name)))
	}

	return dir, nil
}

// Result returns executes the workflow run and returns the result.
func (wr *WorkflowRun) Result(ctx context.Context) (string, error) {
	container, err := wr.run(ctx)
	if err != nil {
		return "", err
	}

	var result Result

	err = container.File("/home/runner/_temp/ghx/result.json").unmarshalContentsToJSON(ctx, &result)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Workflow %s completed with conclusion %s in %s", wr.Config.Workflow, result.Conclusion, result.CompletedAt.Sub(result.StartedAt).String()), nil
}

func (wr *WorkflowRun) run(ctx context.Context) (*Container, error) {
	container, err := wr.container(ctx)
	if err != nil {
		return nil, err
	}

	// loading request scoped configs
	container = container.WithEnvVariable("GHX_WORKFLOW", wr.Config.Workflow)
	container = container.WithEnvVariable("GHX_JOB", wr.Config.Job)
	container = container.WithEnvVariable("GHX_WORKFLOWS_DIR", wr.Config.WorkflowsDir)
	container = container.WithMountedDirectory("/home/runner/_temp/ghx", dag.Directory()).WithEnvVariable("GHX_HOME", "/home/runner/_temp/ghx")

	// execute the workflow
	container = container.WithExec([]string{"/usr/local/bin/ghx"}, ContainerWithExecOpts{ExperimentalPrivilegedNesting: true})

	// unloading request scoped configs
	container = container.WithoutEnvVariable("GHX_WORKFLOW")
	container = container.WithoutEnvVariable("GHX_JOB")
	container = container.WithoutEnvVariable("GHX_WORKFLOWS_DIR")

	return container, nil
}

func (wr *WorkflowRun) container(ctx context.Context) (container *Container, err error) {
	container = dag.Container().From(wr.Config.RunnerImage)

	// set github token as secret
	container = container.WithSecretVariable("GITHUB_TOKEN", wr.Config.Token)

	// load github cli to the container
	container = container.withGithubCli()

	// load ghx to the container
	container, err = container.withGHX(ctx)
	if err != nil {
		return nil, err
	}

	// load repo config to the container
	container, err = container.withRepo(wr.Config.Info, wr.Config.Source)
	if err != nil {
		return nil, err
	}

	// load event config to the container
	container, err = container.withEvent(wr.Config.EventName, wr.Config.EventFile)
	if err != nil {
		return nil, err
	}

	// bind internal services
	container, err = wr.InternalServices.BindServices(ctx, container, InternalServiceOpts{CacheVolumeKeyPrefix: fmt.Sprintf("gale-%s-%s-", wr.Config.Info.Owner, wr.Config.Info.Name)})
	if err != nil {
		return nil, err
	}

	// add env variable to the container to indicate container is configured
	container = container.WithEnvVariable("GALE_CONFIGURED", "true")

	// hacks - TODO: clean-up later
	container = container.WithNewFile("/home/runner/_temp/ghx/secrets/secret.json", ContainerWithNewFileOpts{Contents: "{}"})

	return container, nil
}

// withRepo loads the repository information into the container. context arg and error return value are not used here
// but added to keep the signature of the function consistent with other load*Config functions.
func (c *Container) withRepo(info *GithubRepository, source *Directory) (*Container, error) {
	workdir := filepath.Join("/home", "runner", "work", info.Name, info.Name)

	container := c.WithMountedDirectory(workdir, source).
		WithWorkdir(workdir).
		WithEnvVariable("GITHUB_WORKSPACE", workdir).
		WithEnvVariable("GH_REPO", info.NameWithOwner). // go-gh respects this variable while loading the repository.
		WithEnvVariable("GITHUB_REPOSITORY", info.NameWithOwner).
		WithEnvVariable("GITHUB_REPOSITORY_ID", info.ID).
		WithEnvVariable("GITHUB_REPOSITORY_OWNER", info.Owner.Login).
		WithEnvVariable("GITHUB_REPOSITORY_OWNER_ID", info.Owner.ID).
		WithEnvVariable("GITHUB_REPOSITORY_URL", info.URL)

	return container, nil
}

// withEvent loads the event configuration into the container. context arg and error return value are not used
// here but added to keep the signature of the function consistent with other load*Config functions.
func (c *Container) withEvent(event string, data *File) (*Container, error) {
	container := c

	container = container.WithEnvVariable("GITHUB_EVENT_NAME", event)

	if data != nil {
		path := filepath.Join("/home", "runner", "work", "_temp", "_github_workflow", "event.json")

		container = container.WithMountedFile(path, data).WithEnvVariable("GITHUB_EVENT_PATH", path)
	}

	return container, nil
}
