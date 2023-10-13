package main

func getDefaultConfig() *Config {
	return &Config{
		RepoOpts: &RepoOpts{
			Repo:   "",
			Branch: "",
			Tag:    "",
			Commit: "",
		},
		GithubOpts: &GithubOpts{
			APIURL:     "https://api.github.com",
			GraphqlURL: "https://api.github.com/graphql",
			ServerURL:  "https://github.com",
			Token:      "", // need to check container env for this and then fail if not found.
		},
		EventOpts: &EventOpts{
			EventName: "push",
			EventFile: nil,
		},
	}
}

type Config struct {
	*RepoOpts
	*GithubOpts
	*EventOpts
}

// RepoOpts represents the options for getting repository information.
type RepoOpts struct {
	Repo   string `json:"repo" doc:"name of the repository. name can be in the following formats: OWNER/REPO, HOST/OWNER/REPO, and a full URL. If empty, repository information of the current directory will be used."`
	Branch string `json:"branch" doc:"branch to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch, commit."`
	Tag    string `json:"tag" doc:"tag to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch, commit."`
	Commit string `json:"commit" doc:"commit to load workflows from. Only one of commit, branch or tag can be used. Precedence is as follows: tag, branch, commit."`
}

// GithubOpts represents the options for getting github information.
type GithubOpts struct {
	APIURL     string `doc:"The CloneURL of the Github API." default:"https://api.github.com"`
	GraphqlURL string `doc:"The CloneURL of the Github GraphQL API." default:"https://api.github.com/graphql"`
	ServerURL  string `doc:"The CloneURL of the Github server." default:"https://github.com"`
	Token      string `doc:"GitHub token used for authentication."`
}

// EventOpts represents the options for getting event information.
type EventOpts struct {
	EventName string `doc:"Name of the event that triggered the workflow. e.g. push"`
	EventFile *File  `doc:"The file with the complete webhook event payload."`
}
