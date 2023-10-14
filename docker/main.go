package main

import "context"

// Docker represents the Docker module for Dagger.
type Docker struct{}

// DockerServiceOpts represents the options for binding the Docker module as a service.
type DockerServiceOpts struct {
	CacheVolume *CacheVolume `doc:"The volume to use for caching the Docker data. If not provided, the data is not cached."`
}

// BindAsService binds the Docker module as a service to given container.
func (m *Docker) BindAsService(ctx context.Context, container *Container, opts DockerServiceOpts) (*Container, error) {
	dind := dag.Container().
		From("docker:dind").
		WithUser("root").
		WithEnvVariable("DOCKER_TLS_CERTDIR", ""). // disable TLS
		WithExec([]string{"-H", "tcp://0.0.0.0:2375"}, ContainerWithExecOpts{InsecureRootCapabilities: true}).
		WithExposedPort(2375)

	// If a cache volume is provided, we'll mount it /var/lib/docker.
	if opts.CacheVolume != nil {
		container = container.WithMountedCache("/var/lib/docker", opts.CacheVolume)
	}

	// convert the container to a service.
	service := dind.AsService()

	// get the endpoint of the service to set the DOCKER_HOST environment variable. The reason we're not using the
	// alias for docker is because the service alias is not available in the child containers of the container.
	endpoint, err := service.Endpoint(ctx, ServiceEndpointOpts{Scheme: "tcp"})
	if err != nil {
		return nil, err
	}

	// bind the service to the container and set the DOCKER_HOST environment variable.
	return container.WithServiceBinding("docker", service).WithEnvVariable("DOCKER_HOST", endpoint), nil
}
