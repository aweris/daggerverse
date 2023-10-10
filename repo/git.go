package main

import (
	"fmt"
	"net/url"
	"strings"
)

// original code from github.com/cli/go-gh/v2/internal/git and github.com/cli/go-gh/v2/pkg/repository

const DefaultHost = "github.com"

// repository holds information representing a GitHub repository.
type repository struct {
	Host  string
	Name  string
	Owner string
}

// parse extracts the repository information from the following
// string formats: "OWNER/REPO", "HOST/OWNER/REPO", and a full URL.
// If the format does not specify a host, use the config to determine a host.
func parse(s string) (repository, error) {
	var r repository

	if isURL(s) {
		u, err := parseURL(s)
		if err != nil {
			return r, err
		}

		host, owner, name, err := repoInfoFromURL(u)
		if err != nil {
			return r, err
		}

		r.Host = host
		r.Name = name
		r.Owner = owner

		return r, nil
	}

	parts := strings.SplitN(s, "/", 4)
	for _, p := range parts {
		if len(p) == 0 {
			return r, fmt.Errorf(`expected the "[HOST/]OWNER/REPO" format, got %q`, s)
		}
	}

	switch len(parts) {
	case 3:
		r.Host = parts[0]
		r.Owner = parts[1]
		r.Name = parts[2]
		return r, nil
	case 2:
		r.Host = DefaultHost
		r.Owner = parts[0]
		r.Name = parts[1]
		return r, nil
	default:
		return r, fmt.Errorf(`expected the "[HOST/]OWNER/REPO" format, got %q`, s)
	}
}

func isURL(u string) bool {
	return strings.HasPrefix(u, "git@") || isSupportedProtocol(u)
}

func isSupportedProtocol(u string) bool {
	return strings.HasPrefix(u, "ssh:") ||
		strings.HasPrefix(u, "git+ssh:") ||
		strings.HasPrefix(u, "git:") ||
		strings.HasPrefix(u, "http:") ||
		strings.HasPrefix(u, "git+https:") ||
		strings.HasPrefix(u, "https:")
}

func isPossibleProtocol(u string) bool {
	return isSupportedProtocol(u) ||
		strings.HasPrefix(u, "ftp:") ||
		strings.HasPrefix(u, "ftps:") ||
		strings.HasPrefix(u, "file:")
}

// parseURL normalizes git remote urls.
func parseURL(rawURL string) (u *url.URL, err error) {
	if !isPossibleProtocol(rawURL) &&
		strings.ContainsRune(rawURL, ':') &&
		// Not a Windows path.
		!strings.ContainsRune(rawURL, '\\') {
		// Support scp-like syntax for ssh protocol.
		rawURL = "ssh://" + strings.Replace(rawURL, ":", "/", 1)
	}

	u, err = url.Parse(rawURL)
	if err != nil {
		return
	}

	if u.Scheme == "git+ssh" {
		u.Scheme = "ssh"
	}

	if u.Scheme == "git+https" {
		u.Scheme = "https"
	}

	if u.Scheme != "ssh" {
		return
	}

	if strings.HasPrefix(u.Path, "//") {
		u.Path = strings.TrimPrefix(u.Path, "/")
	}

	if idx := strings.Index(u.Host, ":"); idx >= 0 {
		u.Host = u.Host[0:idx]
	}

	return
}

// repoInfoFromURL extracts GitHub repository information from a git remote URL.
func repoInfoFromURL(u *url.URL) (host string, owner string, name string, err error) {
	if u.Hostname() == "" {
		return "", "", "", fmt.Errorf("no hostname detected")
	}

	parts := strings.SplitN(strings.Trim(u.Path, "/"), "/", 3)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid path: %s", u.Path)
	}

	return normalizeHostname(u.Hostname()), parts[0], strings.TrimSuffix(parts[1], ".git"), nil
}

func normalizeHostname(h string) string {
	return strings.ToLower(strings.TrimPrefix(h, "www."))
}
