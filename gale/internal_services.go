package main

import (
	"context"
	"fmt"
)

type InternalServices struct {
	Dind                bool
	DockerSocket        string
	DockerVolume        *CacheVolume
	ArtifactVolume      *CacheVolume
	ArtifactCacheVolume *CacheVolume
}

func NewInternalServices(opts InternalServiceOpts) *InternalServices {
	return &InternalServices{
		Dind:                opts.Dind,
		DockerSocket:        opts.DockerSocket,
		DockerVolume:        dag.CacheVolume(fmt.Sprintf("%s-docker", opts.CacheVolumeKeyPrefix)),
		ArtifactVolume:      dag.CacheVolume(fmt.Sprintf("%s-artifacts", opts.CacheVolumeKeyPrefix)),
		ArtifactCacheVolume: dag.CacheVolume(fmt.Sprintf("%s-artifactcache", opts.CacheVolumeKeyPrefix)),
	}
}

// BindServices binds the internal services required for running actions
func (is *InternalServices) BindServices(ctx context.Context, c *Container) (container *Container, err error) {
	container = c

	container, err = is.BindDocker(ctx, container, is.DockerVolume)
	if err != nil {
		return nil, err
	}

	container, err = is.BindArtifactService(ctx, container, is.ArtifactVolume)
	if err != nil {
		return nil, err
	}

	container, err = is.BindArtifactCacheService(ctx, container, is.ArtifactCacheVolume)
	if err != nil {
		return nil, err
	}

	return container.WithEnvVariable("ACTIONS_RUNTIME_TOKEN", "token"), nil
}

// BindDocker binds the Docker to the given container.
func (is *InternalServices) BindDocker(_ context.Context, container *Container, volume *CacheVolume) (*Container, error) {
	if !is.Dind {
		socket := dag.Host().UnixSocket(is.DockerSocket)

		return container.WithUnixSocket(is.DockerSocket, socket).
			WithEnvVariable("DOCKER_HOST", fmt.Sprintf("unix://%s", is.DockerSocket)).
			WithMountedCache("/var/lib/docker", volume, ContainerWithMountedCacheOpts{Sharing: Shared}), nil
	}

	return dag.Docker().BindAsService(container, DockerBindAsServiceOpts{CacheVolume: volume}), nil
}

// BindArtifactService binds the Github Actions artifact service to the given container.
func (is *InternalServices) BindArtifactService(ctx context.Context, container *Container, volume *CacheVolume) (*Container, error) {
	service := base().WithEntrypoint([]string{"/usr/local/bin/artifact-service"}).
		WithExposedPort(8080).
		WithEnvVariable("PORT", "8080").
		WithMountedCache("/artifacts", volume, ContainerWithMountedCacheOpts{Sharing: Shared}).
		WithEnvVariable("ARTIFACTS_DIR", "/artifacts").
		AsService()

	endpoint, err := service.Endpoint(ctx, ServiceEndpointOpts{Scheme: "http"})
	if err != nil {
		return nil, err
	}

	// convert the endpoint to expected format for the actions runtime
	endpoint = endpoint + "/"

	return container.WithServiceBinding("artifacts", service).WithEnvVariable("ACTIONS_RUNTIME_URL", endpoint), nil
}

// BindArtifactCacheService binds the Github Actions artifact cache service to the given container.
func (is *InternalServices) BindArtifactCacheService(ctx context.Context, container *Container, volume *CacheVolume) (*Container, error) {
	service := base().WithEntrypoint([]string{"/usr/local/bin/artifactcache-service"}).
		WithExposedPort(8080).
		WithEnvVariable("PORT", "8080").
		WithMountedCache("/cache", volume, ContainerWithMountedCacheOpts{Sharing: Shared}).
		WithEnvVariable("CACHE_DIR", "/cache").
		AsService()

	endpoint, err := service.Endpoint(ctx, ServiceEndpointOpts{Scheme: "http"})
	if err != nil {
		return nil, err
	}

	// convert the endpoint to expected format for the actions runtime
	endpoint = endpoint + "/"

	return container.WithServiceBinding("artifactcache", service).WithEnvVariable("ACTIONS_CACHE_URL", endpoint), nil
}
