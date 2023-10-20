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
		InternalServices: NewInternalServices(
			InternalServiceOpts{
				CacheVolumeKeyPrefix: fmt.Sprintf("gale-%s-%s-", w.Config.Info.Owner, w.Config.Info.Name),
				Dind:                 opts.Dind,
				DockerSocket:         opts.DockerSocket,
			}),
	}
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
func (wr *WorkflowRun) Directory(ctx context.Context, opts WorkflowRunDirectoryOpts) (*Directory, error) {
	container, err := wr.run(ctx)
	if err != nil {
		return nil, err
	}

	runs := container.Directory("/home/runner/_temp/ghx/runs")

	// runs directory should only have one entry with the workflow run id
	entries, err := runs.Entries(ctx)
	if err != nil {
		return nil, err
	}

	wrID := entries[0]

	dir := dag.Directory().WithDirectory("runs", runs)

	if opts.IncludeRepo {
		dir = dir.WithDirectory(fmt.Sprintf("runs/%s/repo", wrID), container.Directory(fmt.Sprintf("/home/runner/work/%s/%s", wr.Config.Info.Name, wr.Config.Info.Name)))
	}

	if opts.IncludeSecrets {
		dir = dir.WithDirectory(fmt.Sprintf("runs/%s/secrets", wrID), container.Directory("/home/runner/_temp/ghx/secrets"))
	}

	if opts.IncludeEvent && wr.Config.EventFile != nil {
		dir = dir.WithFile(fmt.Sprintf("runs/%s/event.json", wrID), container.File(filepath.Join("/home", "runner", "work", "_temp", "_github_workflow", "event.json")))
	}

	if opts.IncludeActions {
		container = container.
			WithExec([]string{"rm", "-rf", "/home/runner/_temp/exported_actions"}).
			WithExec([]string{"cp", "-r", "/home/runner/_temp/ghx/actions", "/home/runner/_temp/exported_actions"})

		dir = dir.WithDirectory(fmt.Sprintf("runs/%s/actions", wrID), container.Directory("/home/runner/_temp/exported_actions"))
	}

	if opts.IncludeArtifacts {
		container = dag.Container().From("alpine:latest").
			WithMountedCache("/artifacts", wr.InternalServices.ArtifactVolume).
			WithExec([]string{"cp", "-r", fmt.Sprintf("/artifacts/%s", wrID), "/exported_artifacts"})

		dir = dir.WithDirectory(fmt.Sprintf("runs/%s/artifacts", wrID), container.Directory("/exported_artifacts"))
	}

	return dir, nil
}

// Result returns executes the workflow run and returns the result.
func (wr *WorkflowRun) Result(ctx context.Context) (string, error) {
	container, err := wr.run(ctx)
	if err != nil {
		return "", err
	}

	var result WorkflowRunResult

	runs := container.Directory("/home/runner/_temp/ghx/runs")

	// runs directory should only have one entry with the workflow run id
	entries, err := runs.Entries(ctx)
	if err != nil {
		return "", err
	}

	wrID := entries[0]

	resultJSON := filepath.Join("/home/runner/_temp/ghx/runs", wrID, "workflow_run.json")

	err = container.File(resultJSON).unmarshalContentsToJSON(ctx, &result)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Workflow %s completed with conclusion %s in %s", result.Name, result.Conclusion, result.Duration), nil
}

func (wr *WorkflowRun) run(ctx context.Context) (*Container, error) {
	container, err := wr.container(ctx)
	if err != nil {
		return nil, err
	}

	// loading request scoped configs

	var (
		metadataCache = dag.CacheVolume("gale-metadata")
		actionsCache  = dag.CacheVolume(fmt.Sprintf("gale-%s-%s-actions", wr.Config.Info.Owner, wr.Config.Info.Name))
	)

	if wr.Config.Debug {
		container = container.WithEnvVariable("RUNNER_DEBUG", "1")
	}

	container = container.WithEnvVariable("GHX_WORKFLOW", wr.Config.Workflow)
	container = container.WithEnvVariable("GHX_JOB", wr.Config.Job)
	container = container.WithEnvVariable("GHX_WORKFLOWS_DIR", wr.Config.WorkflowsDir)
	container = container.WithEnvVariable("GHX_HOME", "/home/runner/_temp/ghx").
		WithMountedDirectory("/home/runner/_temp/ghx", dag.Directory()).
		WithMountedCache("/home/runner/_temp/ghx/metadata", metadataCache, ContainerWithMountedCacheOpts{Sharing: Shared}).
		WithMountedCache("/home/runner/_temp/ghx/actions", actionsCache, ContainerWithMountedCacheOpts{Sharing: Shared})

	// workaround for disabling cache
	container = container.WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano))

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
	container, err = container.withRepo(wr.Config.Info, wr.Config.Source, wr.Config.Ref)
	if err != nil {
		return nil, err
	}

	// load event config to the container
	container, err = container.withEvent(wr.Config.EventName, wr.Config.EventFile)
	if err != nil {
		return nil, err
	}

	// bind internal services
	container, err = wr.InternalServices.BindServices(ctx, container)
	if err != nil {
		return nil, err
	}

	// add env variable to the container to indicate container is configured
	container = container.WithEnvVariable("GALE_CONFIGURED", "true")

	return container, nil
}

// withRepo loads the repository information into the container. context arg and error return value are not used here
// but added to keep the signature of the function consistent with other load*Config functions.
func (c *Container) withRepo(info *GithubRepository, source *Directory, ref *RepositoryRef) (*Container, error) {
	workdir := filepath.Join("/home", "runner", "work", info.Name, info.Name)

	container := c.WithMountedDirectory(workdir, source).
		WithWorkdir(workdir).
		WithEnvVariable("GITHUB_WORKSPACE", workdir).
		WithEnvVariable("GH_REPO", info.NameWithOwner). // go-gh respects this variable while loading the repository.
		WithEnvVariable("GITHUB_REPOSITORY", info.NameWithOwner).
		WithEnvVariable("GITHUB_REPOSITORY_OWNER", info.Owner).
		WithEnvVariable("GITHUB_REPOSITORY_URL", info.URL).
		WithEnvVariable("GITHUB_REF", ref.Ref).
		WithEnvVariable("GITHUB_REF_NAME", ref.RefName).
		WithEnvVariable("GITHUB_REF_TYPE", ref.RefType).
		WithEnvVariable("GITHUB_SHA", ref.SHA)

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
