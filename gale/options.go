package main

// RepoOpts represents the options for getting repository information.
type RepoOpts struct {
	Repo         string `json:"repo" doc:"owner/repo to load workflows from. If empty, repository information of the current directory will be used."`
	WorkflowsDir string `json:"workflows-dir" doc:"directory to load workflows from. If empty, workflows will be loaded from the default directory." default:".github/workflows"`
	Branch       string `json:"branch" doc:"branch to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch."`
	Tag          string `json:"tag" doc:"tag to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch."`
}

func (o *RepoOpts) toExecArgs() []string {
	var args []string

	if o.Repo != "" {
		args = append(args, "--repo", o.Repo)
	}

	if o.WorkflowsDir != "" {
		args = append(args, "--workflows-dir", o.WorkflowsDir)
	}

	if o.Branch != "" {
		args = append(args, "--branch", o.Branch)
	}

	if o.Tag != "" {
		args = append(args, "--tag", o.Tag)
	}

	return args
}

// RunOpts represents the options for running a workflow.
type RunOpts struct {
	Job string `json:"job" doc:"name of the job to run. If empty, all jobs in the workflow will be run."`
}

func (o *RunOpts) toExecArgs() []string {
	var args []string

	if o.Job != "" {
		args = append(args, "--job", o.Job)
	}

	return args
}

// RunnerOpts represents the options for runner related configurations to be set on container.
type RunnerOpts struct {
	Debug bool `json:"debug" doc:"debug mode. If true, the workflow will be run in debug mode."`
}
