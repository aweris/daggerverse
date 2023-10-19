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
	args := []string{"repo", "view", opts.Repo, "--json", "id,name,owner,nameWithOwner,url,defaultBranchRef"}

	container := gh().WithSecretVariable("GITHUB_TOKEN", g.Config.Token)

	// if the repository is not set, mount the current directory as the repository
	if opts.Repo == "" {
		container = container.WithMountedDirectory("/src", opts.Source).WithWorkdir("/src")
	}

	var info *GithubRepository

	err := container.WithExec(args).asJSON(ctx, &info)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get repository information", err)
	}

	return info, nil
}

// getGithubRepositorySource returns a new Directory with given repository information and options. If all options are
// empty, the current directory is returned. If repo name is provided but no other options are provided, the default
// branch is used.
func (g *Gale) getGithubRepositorySource(info *GithubRepository, opts RepoOpts) (*Directory, error) {
	source := opts.Source

	// if all options are empty, use the current directory
	if opts.Repo == "" && opts.Branch == "" && opts.Commit == "" && opts.Tag == "" {
		return dag.Host().Directory("."), nil
	}

	// if all options are empty, find default branch and use that
	if opts.Tag == "" && opts.Branch == "" && opts.Commit == "" {
		opts.Branch = info.DefaultBranchRef.Name
	}

	switch {
	case opts.Tag != "":
		source = dag.Git(info.URL, GitOpts{KeepGitDir: true}).Tag(opts.Tag).Tree()
	case opts.Branch != "":
		source = dag.Git(info.URL, GitOpts{KeepGitDir: true}).Branch(opts.Branch).Tree()
	case opts.Commit != "":
		source = dag.Git(info.URL, GitOpts{KeepGitDir: true}).Commit(opts.Commit).Tree()
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
	case isRemote:
		ref = fmt.Sprintf("refs/heads/%s", info.DefaultBranchRef.Name)
		refName = info.DefaultBranchRef.Name
		refType = "branch"
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
