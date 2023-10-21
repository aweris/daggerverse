package main

// options - structs exposed to dagger as options

// RepoOpts represents the options for getting repository information.
type RepoOpts struct {
	Repo   string     `json:"repo" doc:"name of the repository. name can be in the following formats: OWNER/REPO, HOST/OWNER/REPO, and a full URL. If empty, the source will be mounted as the repository."`
	Source *Directory `json:"-" doc:"the source to load the repository from. If additional commit, branch or tag options are provided, the source will be cloned and the commit, branch or tag will be checked out."`
	Branch string     `json:"branch" doc:"branch to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch, commit."`
	Tag    string     `json:"tag" doc:"tag to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch, commit."`
	Commit string     `json:"commit" doc:"commit to load workflows from. Only one of commit, branch or tag can be used. Precedence is as follows: tag, branch, commit."`
}

// WorkflowOpts represents the options for getting workflow information.
type WorkflowOpts struct {
	WorkflowsDir string `doc:"The directory to look for workflows." default:".github/workflows"`
}

// WorkflowRunOpts represents the options for running a workflow.
type WorkflowRunOpts struct {
	Workflow     string  `doc:"The workflow to run."`
	Job          string  `doc:"The job name to run. If empty, all jobs will be run."`
	EventName    string  `doc:"Name of the event that triggered the workflow. e.g. push"`
	EventFile    *File   `doc:"The file with the complete webhook event payload."`
	RunnerImage  string  `doc:"The image to use for the runner." default:"ghcr.io/catthehacker/ubuntu:act-latest"`
	Debug        bool    `doc:"Enable debug mode." default:"false"`
	Dind         bool    `doc:"Enable Docker-in-Dagger mode. This will start a Docker Service in dagger and bind it to the container. instead of using the host's Docker socket." default:"false"`
	DockerSocket *Socket `doc:"Docker socket to enable Docker-out-of-Dagger mode. This will bind the Docker socket to the container. This is only used if DinD is false."`
	Token        string  `doc:"The GitHub token to use for authentication."`
}

// WorkflowRunDirectoryOpts represents the options for exporting a workflow run.
type WorkflowRunDirectoryOpts struct {
	IncludeRepo      bool `doc:"Include the repository source in the exported directory." default:"false"`
	IncludeSecrets   bool `doc:"Include the secrets in the exported directory." default:"false"`
	IncludeEvent     bool `doc:"Include the event file in the exported directory." default:"false"`
	IncludeActions   bool `doc:"Include the custom action repo in the exported directory." default:"false"`
	IncludeArtifacts bool `doc:"Include the artifacts in the exported directory." default:"false"`
}

// InternalServiceOpts represents the options for internal services.
type InternalServiceOpts struct {
	CacheVolumeKeyPrefix string  `doc:"The prefix to use for the cache volume key." default:"gale"`
	Dind                 bool    `doc:"Enable Docker-in-Dagger mode. This will start a Docker Service in dagger and bind it to the container. instead of using the host's Docker socket." default:"false"`
	DockerSocket         *Socket `doc:"Docker socket to enable Docker-out-of-Dagger mode. This will bind the Docker socket to the container. This is only used if DinD is false."`
}

// models - internal structs used by gale

// GithubRepository represents a GitHub repository
type GithubRepository struct {
	Name          string `json:"name"`
	NameWithOwner string `json:"nameWithOwner"`
	URL           string `json:"url" `
	Owner         string `json:"owner"`
}

// RepositoryRef represents a git repository ref information. This is used to get the ref information from the local git
// from the repository source.
type RepositoryRef struct {
	Ref      string // Ref is the branch or tag ref that triggered the workflow
	RefName  string // RefName is the short name (without refs/heads/ prefix) of the branch or tag ref that triggered the workflow.
	RefType  string // RefType is the type of ref that triggered the workflow. Possible values are branch, tag, or empty, if neither
	SHA      string // SHA is the commit SHA that triggered the workflow. The value of this commit SHA depends on the event that
	ShortSHA string // ShortSHA is the short commit SHA that triggered the workflow. The value of this commit SHA depends on the event that
	IsRemote bool   // IsRemote is true if the ref is a remote ref.
}

// WorkflowRunResult represents the result of a workflow run.
type WorkflowRunResult struct {
	Ran        bool   `json:"ran"`        // Ran indicates if the execution ran
	Path       string `json:"path"`       // Path is the path to the workflow run directory
	Name       string `json:"name"`       // Name is the name of the workflow run
	Conclusion string `json:"conclusion"` // Conclusion of the execution
	Duration   string `json:"duration"`   // Duration of the execution
}

// configurations - configurations used by gale entrypoint to easier configuration composition and reuse

// RepoConfig holds the repository information and source.
type RepoConfig struct {
	Info   *GithubRepository
	Source *Directory
	Ref    *RepositoryRef
}

type WorkflowsConfig struct {
	*RepoConfig

	WorkflowsDir string
}

type WorkflowRunConfig struct {
	*WorkflowsConfig
	*WorkflowRunOpts
}
