package main

// options - structs exposed to dagger as options

// RepoOpts represents the options for getting repository information.
type RepoOpts struct {
	Repo   string `json:"repo" doc:"name of the repository. name can be in the following formats: OWNER/REPO, HOST/OWNER/REPO, and a full URL. If empty, repository information of the current directory will be used."`
	Branch string `json:"branch" doc:"branch to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch, commit."`
	Tag    string `json:"tag" doc:"tag to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch, commit."`
	Commit string `json:"commit" doc:"commit to load workflows from. Only one of commit, branch or tag can be used. Precedence is as follows: tag, branch, commit."`
}

// WorkflowOpts represents the options for getting workflow information.
type WorkflowOpts struct {
	WorkflowsDir string `doc:"The directory to look for workflows." default:".github/workflows"`
}

// WorkflowRunOpts represents the options for running a workflow.
type WorkflowRunOpts struct {
	Workflow    string `doc:"The workflow to run."`
	Job         string `doc:"The job name to run. If empty, all jobs will be run."`
	EventName   string `doc:"Name of the event that triggered the workflow. e.g. push"`
	EventFile   *File  `doc:"The file with the complete webhook event payload."`
	RunnerImage string `doc:"The image to use for the runner." default:"ghcr.io/aweris/gale/runner/ubuntu:22.04"`
}

// WorkflowRunDirectoryOpts represents the options for exporting a workflow run.
type WorkflowRunDirectoryOpts struct {
	IncludeRepo     bool `doc:"Include the repository source in the exported directory." default:"false"`
	IncludeMetadata bool `doc:"Include the workflow run metadata in the exported directory." default:"false"`
	IncludeSecrets  bool `doc:"Include the secrets in the exported directory." default:"false"`
	IncludeEvent    bool `doc:"Include the event file in the exported directory." default:"false"`
	IncludeActions  bool `doc:"Include the custom action repo in the exported directory." default:"false"`
}

// models - internal structs used by gale

// GithubRepository represents a GitHub repository
type GithubRepository struct {
	ID               string                    `json:"id"`
	Name             string                    `json:"name"`
	NameWithOwner    string                    `json:"nameWithOwner"`
	URL              string                    `json:"url" `
	Owner            GithubRepositoryOwner     `json:"owner"`
	DefaultBranchRef GithubRepositoryBranchRef `json:"defaultBranchRef"`
}

// GithubRepositoryOwner represents a GitHub repository owner
type GithubRepositoryOwner struct {
	ID    string `json:"id"`
	Login string `json:"login"`
}

// GithubRepositoryBranchRef represents a GitHub repository branch ref
type GithubRepositoryBranchRef struct {
	Name string `json:"name"`
}

// configurations - configurations used by gale entrypoint to easier configuration composition and reuse

// AuthConfig holds the GITHUB_TOKEN secret.
type AuthConfig struct {
	Token *Secret
}

// RepoConfig holds the repository information and source.
type RepoConfig struct {
	*AuthConfig

	Info   *GithubRepository
	Source *Directory
}

type WorkflowsConfig struct {
	*RepoConfig

	WorkflowsDir string
}

type WorkflowRunConfig struct {
	*WorkflowsConfig
	*WorkflowRunOpts
}
