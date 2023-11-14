package main

import "context"

// Docker represents the Docker module for Dagger.
type Docker struct {
	CacheVolume *CacheVolume
}

// WithCacheVolume sets the cache volume for the Docker module.
func (m *Docker) WithCacheVolume(key string) *Docker {
	m.CacheVolume = dag.CacheVolume(key)
	return m
}

// Dind returns docker:dind as a service.
func (m *Docker) Dind() *Service {
	dind := dag.Container().
		From("docker:dind").
		WithUser("root").
		WithEnvVariable("DOCKER_TLS_CERTDIR", ""). // disable TLS
		WithExec([]string{"-H", "tcp://0.0.0.0:2375"}, ContainerWithExecOpts{InsecureRootCapabilities: true}).
		WithExposedPort(2375)

	// If a cache volume is provided, we'll mount it /var/lib/docker.
	if m.CacheVolume != nil {
		dind = dind.WithMountedCache("/var/lib/docker", m.CacheVolume, ContainerWithMountedCacheOpts{Sharing: Shared})
	}

	return dind.AsService()
}

// BindAsService binds the Docker module as a service to given container.
func (m *Docker) BindAsService(
	ctx context.Context,
	// container to bind the docker service to
	container *Container,
) (*Container, error) {
	// convert the container to a service.
	service := m.Dind()

	// get the endpoint of the service to set the DOCKER_HOST environment variable. The reason we're not using the
	// alias for docker is because the service alias is not available in the child containers of the container.
	endpoint, err := service.Endpoint(ctx, ServiceEndpointOpts{Scheme: "tcp"})
	if err != nil {
		return nil, err
	}

	// bind the service to the container and set the DOCKER_HOST environment variable.
	return container.WithServiceBinding("docker", service).WithEnvVariable("DOCKER_HOST", endpoint), nil
}
