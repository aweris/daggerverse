// Runs a k3s server than can be accessed both locally and in your pipelines

package main

import (
	"fmt"
	"strings"
	"time"

	"dagger/k-3-s/internal/dagger"
)

type K3S struct {
	// +private
	Ctr *Container

	// +private
	Cache *CacheVolume

	// +private
	HttpListenPort int
}

func New(
	// Name of the k3s cluster
	//
	// +optional
	// +default="default"
	name string,

	// Override the base rancher/k3s container with a custom one
	//
	// +optional
	base *Container,

	// HTTPS listen port
	//
	// +optional
	// +default=6443
	port int,
) *K3S {

	if base == nil {
		base = dag.Container().From("rancher/k3s")
	}

	return &K3S{
		Cache:          dag.CacheVolume("k3s_config_" + name),
		Ctr:            base,
		HttpListenPort: port,
	}
}

// Returns a configured container for the k3s
func (m *K3S) Container() *Container {
	return m.Ctr.
		With(m.entrypoint).
		WithMountedCache("/etc/rancher/k3s", m.Cache).
		WithMountedTemp("/etc/lib/cni").
		WithMountedTemp("/var/lib/kubelet").
		WithMountedTemp("/var/lib/rancher/k3s").
		WithMountedTemp("/var/log").
		WithExposedPort(m.HttpListenPort)
}

// Returns initialized k3s cluster
func (m *K3S) Server() *Service {
	return m.Container().With(m.k3sServer).AsService()
}

// Returns the kubeconfig file from the k3s container
func (m *K3S) Kubeconfig(
	// Indicates that the kubeconfig should be use localhost instead of the container IP. This is useful when running k3s as service
	//
	// +optional
	// +default=false
	local bool,
) *File {
	return dag.Container().
		From("alpine").
		With(m.cachebuster). // cache buster to force the copy of the k3s.yaml
		WithMountedCache("/cache/k3s", m.Cache).
		WithExec([]string{"cp", "/cache/k3s/k3s.yaml", "k3s.yaml"}).
		With(func(ctr *Container) *Container {
			if !local {
				return ctr
			}

			return ctr.WithExec([]string{"sed", "-i", `s/https:.*:/https:\/\/localhost:/g`, "k3s.yaml"})
		}).
		File("k3s.yaml")
}

// Helper functions used to configure the k3s container

// a helper function to add the entrypoint to the container
func (_ *K3S) entrypoint(ctr *Container) *Container {
	var (
		file = dag.CurrentModule().Source().File("hack/entrypoint.sh")
		opts = dagger.ContainerWithFileOpts{Permissions: 0o755}
	)

	return ctr.WithFile("/usr/bin/entrypoint.sh", file, opts).WithEntrypoint([]string{"entrypoint.sh"})
}

// helper function configure the k3s server command execution
func (m *K3S) k3sServer(ctr *Container) *Container {
	// k3s server -- options
	opts := []string{"k3s", "server"}

	opts = append(opts, "--https-listen-port", fmt.Sprintf("%d", m.HttpListenPort))
	opts = append(opts, "--disable", "traefik")
	opts = append(opts, "--disable", "metrics-server")

	return ctr.WithExec([]string{"sh", "-c", strings.Join(opts, " ")}, ContainerWithExecOpts{InsecureRootCapabilities: true})
}

// helper function to add a cache buster to the container. This will force
// the container execute follow-up steps instead of using the cache
func (_ *K3S) cachebuster(ctr *Container) *Container {
	return ctr.WithEnvVariable("CACHE_BUSTER", time.Now().String())
}
