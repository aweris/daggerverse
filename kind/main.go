package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Kind represents the KinD module for Dagger.
type Kind struct{}

// Exec executes the kind command with the given arguments. If no arguments are given, --help is used.
func (m *Kind) Exec(_ context.Context, args []string) *Container {
	// If no arguments are given, we'll default to --help.
	if len(args) == 0 {
		args = []string{"--help"}
	}

	return exec(args)
}

// Cluster returns the cluster with the given name. If no name is given, the default name, kind, is used. If a cluster
// already exists with the given name, it marks the cluster as existing to avoid creating it again.
func (m *Kind) Cluster(ctx context.Context, name string) (*Cluster, error) {
	if name == "" {
		name = "kind"
	}

	clusters, err := m.Exec(ctx, []string{"get", "clusters"}).Stdout(ctx)
	if err != nil {
		return nil, err
	}

	exist := false

	for _, cluster := range strings.Split(clusters, "\n") {
		if cluster == name {
			fmt.Printf("Cluster %s already exists.\n", name)
			exist = true
			break
		}
	}

	return &Cluster{Name: name, Exists: exist}, nil
}

// Cluster represents a KinD cluster.
type Cluster struct {
	Name   string // Name is the name of the cluster.
	Exists bool   // Exists is true if the cluster exists on the host with the given name.
}

// Create creates the cluster if it doesn't already exist.
func (m *Cluster) Create(ctx context.Context) error {
	if m.Exists {
		fmt.Printf("Cluster %s already exists.\n", m.Name)
		return nil
	}

	_, err := exec([]string{"create", "cluster", "--name", m.Name}).Sync(ctx)
	if err != nil {
		return err
	}

	m.Exists = true

	return nil
}

// Kubeconfig returns the kubeconfig file for the cluster. If internal is true, the internal address is used. Otherwise,
// the external address is used in the kubeconfig. Internal config is useful for running kubectl commands from within
// other containers.
func (m *Cluster) Kubeconfig(_ context.Context, internal bool) *File {
	cmd := []string{"export", "kubeconfig", "--name", m.Name}

	if internal {
		cmd = append(cmd, "--internal")
	}

	return exec(cmd).Directory("/root/.kube").File("config")
}

// Logs returns the directory containing the cluster logs.
func (m *Cluster) Logs(_ context.Context) *Directory {
	dir := filepath.Join("tmp", m.Name, "logs")

	return exec([]string{"export", "logs", dir, "--name", m.Name}).Directory(dir)
}

// Delete deletes the cluster if it exists.
func (m *Cluster) Delete(ctx context.Context) error {
	if !m.Exists {
		fmt.Printf("Cluster %s doesn't exist.\n", m.Name)
		return nil
	}

	_, err := exec([]string{"delete", "cluster", "--name", m.Name}).Sync(ctx)
	if err != nil {
		return err
	}

	m.Exists = false

	return nil
}

// exec executes the kind command with the given arguments.
func exec(args []string) *Container {
	return container().WithExec(append([]string{"kind"}, args...))
}

// container returns a container with the docker and kind binaries installed and the docker socket mounted. As last
// step, it adds a CACHE_BUSTER environment variable to the container to avoid using the cache when running the
// commands.
func container() *Container {
	socket := "/var/run/docker.sock"

	if env := os.Getenv("DOCKER_HOST"); env != "" {
		socket = strings.TrimPrefix(env, "unix://")
	}

	return dag.Container().
		From("alpine:latest").
		WithUser("root").
		WithExec([]string{"apk", "add", "--no-cache", "docker", "kind"}).
		WithUnixSocket("/var/run/docker.sock", dag.Host().UnixSocket(socket)).
		WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano))
}
