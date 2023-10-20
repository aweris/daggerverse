package main

import (
	"bufio"
	"context"
	"fmt"
	"path"
	"strings"
)

// getGithubRepository returns a new GithubRepository with given repository name and options.
func (g *Gale) getGithubRepository(ctx context.Context, opts RepoOpts) (*GithubRepository, error) {
	container := git().WithMountedDirectory("/src", opts.Source).WithWorkdir("/src")

	out, err := container.WithExec([]string{"config", "--get", "remote.origin.url"}).Stdout(ctx)
	if err != nil {
		return nil, err
	}

	url := strings.TrimSpace(out)

	owner, name, err := parseGithubURL(url)
	if err != nil {
		return nil, err
	}

	return &GithubRepository{
		Owner:         owner,
		Name:          name,
		NameWithOwner: fmt.Sprintf("%s/%s", owner, name),
		URL:           url,
	}, nil

}

// getGithubRepositorySource returns a new Directory with given repository information and options. If all options are
// empty, the current directory is returned. If repo name is provided but no other options are provided, the default
// branch is used.
func (g *Gale) getGithubRepositorySource(opts RepoOpts) (*Directory, error) {
	source := opts.Source

	// if all options are empty, use the current directory
	if opts.Repo == "" && opts.Branch == "" && opts.Commit == "" && opts.Tag == "" {
		return source, nil
	}

	url := fmt.Sprintf("https://github.com/%s.git", opts.Repo)

	switch {
	case opts.Tag != "":
		source = dag.Git(url, GitOpts{KeepGitDir: true}).Tag(opts.Tag).Tree()
	case opts.Branch != "":
		source = dag.Git(url, GitOpts{KeepGitDir: true}).Branch(opts.Branch).Tree()
	case opts.Commit != "":
		source = dag.Git(url, GitOpts{KeepGitDir: true}).Commit(opts.Commit).Tree()
	default:
		return nil, fmt.Errorf("couldn't find a repository to load") // this should never happen, added for defensive programming
	}

	return source, nil
}

// getRepositoryRef returns a new RepositoryRef with given repository information and options.
func (g *Gale) getRepositoryRef(ctx context.Context, info *GithubRepository, source *Directory, opts RepoOpts) (*RepositoryRef, error) {
	out, err := git().WithMountedDirectory("/src", source).WithWorkdir("/src").
		WithExec([]string{"rev-parse", "HEAD"}).
		Stdout(ctx)
	if err != nil {
		return nil, err
	}

	var (
		ref      = ""
		refName  = ""
		refType  = ""
		head     = strings.TrimSpace(out)
		isRemote = opts.Repo != "" || opts.Branch != "" || opts.Tag != "" || opts.Commit != ""
	)

	switch {
	case opts.Tag != "":
		ref = fmt.Sprintf("refs/tags/%s", opts.Tag)
		refName = opts.Tag
		refType = "tag"
	case opts.Branch != "":
		ref = fmt.Sprintf("refs/heads/%s", opts.Branch)
		refName = opts.Branch
		refType = "branch"
	case opts.Commit != "":
		ref = fmt.Sprintf("refs/heads/%s", opts.Commit)
		refName = opts.Commit
		refType = "commit"
	default:
		ref, err = getRefFromDirectory(ctx, head, source)
		if err != nil {
			return nil, err
		}

		refName = path.Base(ref)

		switch {
		case strings.HasPrefix(ref, "refs/tags/"):
			refType = "tag"
		case strings.HasPrefix(ref, "refs/heads/"):
			refType = "branch"
		default:
			refType = "commit"
		}

	}
	return &RepositoryRef{
		Ref:      ref,
		RefName:  refName,
		RefType:  refType,
		SHA:      head,
		ShortSHA: head[:7],
		IsRemote: isRemote,
	}, nil
}

func getRefFromDirectory(ctx context.Context, head string, source *Directory) (string, error) {
	out, err := git().WithMountedDirectory("/src", source).WithWorkdir("/src").
		WithExec([]string{"show-ref"}).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(out))

	found := ""
	for scanner.Scan() {
		ref := scanner.Text()

		parts := strings.Fields(ref)

		if len(parts) < 2 {
			continue
		}

		ref = strings.TrimSpace(parts[0])

		if ref == head {
			found = strings.TrimSpace(parts[1])
			break
		}
	}

	if found == "" {
		return "", fmt.Errorf("no ref found for %s", head)
	}

	return found, nil
}

func parseGithubURL(url string) (owner, repo string, err error) {
	if strings.HasPrefix(url, "git@github.com:") {
		url = strings.TrimPrefix(url, "git@github.com:")
		parts := strings.Split(url, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid SSH GitHub URL")
		}
		owner = parts[0]
		repo = strings.TrimSuffix(parts[1], ".git")
	} else if strings.HasPrefix(url, "https://github.com/") {
		url = strings.TrimPrefix(url, "https://github.com/")
		parts := strings.Split(url, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid HTTPS GitHub URL")
		}
		owner = parts[0]
		repo = strings.TrimSuffix(parts[1], ".git")
	} else {
		return "", "", fmt.Errorf("invalid GitHub URL")
	}
	return owner, repo, nil
}
