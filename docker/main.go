package main

import "context"

// Docker represents the Docker module for Dagger.
type Docker struct {
	CacheVolume *CacheVolume
}

// WithCacheVolume sets the cache volume with the given name to be mounted to /var/lib/docker in the service container.
func (m *Docker) WithCacheVolume(name string) *Docker {
	m.CacheVolume = dag.CacheVolume(name)
	return m
}

// Service returns the Docker service. Set DOCKER_HOST to the service endpoint to use it.
func (m *Docker) Service(_ context.Context) *Service {
	container := dag.Container().
		From("docker:dind").
		WithUser("root").
		WithEnvVariable("DOCKER_TLS_CERTDIR", ""). // disable TLS
		WithExec([]string{"-H", "tcp://0.0.0.0:2375"}, ContainerWithExecOpts{InsecureRootCapabilities: true}).
		WithExposedPort(2375)

	if m.CacheVolume != nil {
		container = container.WithMountedCache("/var/lib/docker", m.CacheVolume)
	}

	return container.Service()
}
