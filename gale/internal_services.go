package main

import (
	"context"
	"fmt"
)

type InternalServices struct {
	Artifact            *Container
	ArtifactCache       *Container
	DockerVolume        *CacheVolume
	ArtifactVolume      *CacheVolume
	ArtifactCacheVolume *CacheVolume
}

type InternalServiceOpts struct {
	CacheVolumeKeyPrefix string `doc:"The prefix to use for the cache volume key." default:"gale"`
}

func NewInternalServices(opts InternalServiceOpts) *InternalServices {
	return &InternalServices{
		DockerVolume:        dag.CacheVolume(fmt.Sprintf("%s-docker", opts.CacheVolumeKeyPrefix)),
		ArtifactVolume:      dag.CacheVolume(fmt.Sprintf("%s-artifacts", opts.CacheVolumeKeyPrefix)),
		ArtifactCacheVolume: dag.CacheVolume(fmt.Sprintf("%s-artifactcache", opts.CacheVolumeKeyPrefix)),
	}
}

// BindServices binds the internal services required for running actions
func (is *InternalServices) BindServices(ctx context.Context, c *Container) (container *Container, err error) {
	container = c

	container, err = is.BindDockerService(ctx, container, is.DockerVolume)
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

// BindDockerService binds the Docker service to the given container.
func (is *InternalServices) BindDockerService(_ context.Context, container *Container, volume *CacheVolume) (*Container, error) {
	return dag.Docker().BindAsService(container, DockerBindAsServiceOpts{CacheVolume: volume}), nil
}

// BindArtifactService binds the Github Actions artifact service to the given container.
func (is *InternalServices) BindArtifactService(ctx context.Context, container *Container, volume *CacheVolume) (*Container, error) {
	is.Artifact = base().WithEntrypoint([]string{"/usr/local/bin/artifact-service"}).
		WithExposedPort(8080).
		WithEnvVariable("PORT", "8080").
		WithMountedCache("/artifacts", volume, ContainerWithMountedCacheOpts{Sharing: Shared}).
		WithEnvVariable("ARTIFACTS_DIR", "/artifacts")

	service := is.Artifact.AsService()

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
	is.ArtifactCache = base().WithEntrypoint([]string{"/usr/local/bin/artifactcache-service"}).
		WithExposedPort(8080).
		WithEnvVariable("PORT", "8080").
		WithMountedCache("/cache", volume, ContainerWithMountedCacheOpts{Sharing: Shared}).
		WithEnvVariable("CACHE_DIR", "/cache")

	service := is.ArtifactCache.AsService()

	endpoint, err := service.Endpoint(ctx, ServiceEndpointOpts{Scheme: "http"})
	if err != nil {
		return nil, err
	}

	// convert the endpoint to expected format for the actions runtime
	endpoint = endpoint + "/"

	return container.WithServiceBinding("artifactcache", service).WithEnvVariable("ACTIONS_CACHE_URL", endpoint), nil
}
