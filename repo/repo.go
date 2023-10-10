package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
)

// Repo is a module for loading repository information from GitHub.
type Repo struct{}

// RepoOpts represents the options for getting repository information.
type RepoOpts struct {
	Repo   string `json:"repo" doc:"name of the repository. name can be in the following formats: OWNER/REPO, HOST/OWNER/REPO, and a full URL. If empty, repository information of the current directory will be used."`
	Branch string `json:"branch" doc:"branch to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch, commit."`
	Tag    string `json:"tag" doc:"tag to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch, commit."`
	Commit string `json:"commit" doc:"commit to load workflows from. Only one of commit, branch or tag can be used. Precedence is as follows: tag, branch, commit."`
}

// RepoInfo represents the repository information and the source directory.
type RepoInfo struct {
	Owner    string     `json:"owner"`    // Owner is the owner of the repository.
	Name     string     `json:"name"`     // Name is the name of the repository.
	CloneURL string     `json:"cloneUrl"` // CloneURL is the clone url of the repository. Format: https://HOST/OWNER/REPO.git
	Source   *Directory `json:"source"`   // Source is the Source directory of the repository.
}

func (r *RepoInfo) FullName() string {
	return fmt.Sprintf("%s/%s", r.Owner, r.Name)
}

// Load loads the repository information from the specified options. If repo option is empty, the current directory
// will be used as the repository.
func (_ *Repo) Load(ctx context.Context, opts RepoOpts) (*RepoInfo, error) {
	var current *Directory

	// if repo option is empty, use the current directories remote origin url as repo option
	if opts.Repo == "" {
		current = dag.Host().Directory(".")

		remote, err := getRemoteFromDirectory(ctx, current)
		if err != nil {
			return nil, err
		}

		// set repo option to the remote origin url
		opts.Repo = remote
	}

	info, err := parseRepoInfo(opts.Repo)
	if err != nil {
		return nil, err
	}

	// if all options are empty, find default branch and use that
	if opts.Tag == "" && opts.Branch == "" && opts.Commit == "" {
		branch, err := getDefaultBranchFromRemote(ctx, info.CloneURL)
		if err != nil {
			return nil, err
		}

		opts.Branch = branch
	}

	switch {
	case opts.Tag != "":
		info.Source = dag.Git(info.CloneURL, GitOpts{KeepGitDir: true}).Tag(opts.Tag).Tree()
	case opts.Branch != "":
		info.Source = dag.Git(info.CloneURL, GitOpts{KeepGitDir: true}).Branch(opts.Branch).Tree()
	case opts.Commit != "":
		info.Source = dag.Git(info.CloneURL, GitOpts{KeepGitDir: true}).Commit(opts.Commit).Tree()
	case current != nil:
		info.Source = current
	default:
		return nil, fmt.Errorf("couldn't find a repository to load")
	}

	return info, nil
}

// getDefaultBranchFromRemote returns the default branch of the specified remote url.
func getDefaultBranchFromRemote(ctx context.Context, url string) (string, error) {
	out, err := git().WithExec([]string{"git", "ls-remote", "--symref", url, "HEAD"}).Stdout(ctx)
	if err != nil {
		return "", err
	}

	out = strings.TrimSpace(out)

	scanner := bufio.NewScanner(bytes.NewBufferString(out))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())

		if len(fields) < 3 {
			continue
		}

		if fields[0] == "ref:" && fields[2] == "HEAD" {
			return strings.TrimPrefix(fields[1], "refs/heads/"), nil
		}
	}

	return "", fmt.Errorf("could not find default branch for %s", url)
}

// parseRepoInfo parses the repo option and returns the repository information without any source.
func parseRepoInfo(s string) (*RepoInfo, error) {
	// parse repo option
	repo, err := parse(s)
	if err != nil {
		return nil, err
	}

	info := &RepoInfo{
		Owner:    repo.Owner,
		Name:     repo.Name,
		CloneURL: fmt.Sprintf("https://%s/%s/%s.git", repo.Host, repo.Owner, repo.Name),
	}

	return info, nil
}

// getRemoteFromDirectory returns the remote origin url of the git repository in the specified directory.
func getRemoteFromDirectory(ctx context.Context, dir *Directory) (string, error) {
	out, err := git().
		WithMountedDirectory("/src", dir).
		WithWorkdir("/src").
		WithExec([]string{"git", "config", "--get", "remote.origin.url"}).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

// git returns a dagger container with git package installed.
func git() *Container {
	return dag.Apko().Wolfi([]string{"git"})
}
