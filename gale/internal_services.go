package main

import (
	"context"
	"fmt"
)

type InternalServices struct {
	Artifact      *Container
	ArtifactCache *Container
}

type InternalServiceOpts struct {
	CacheVolumeKeyPrefix string `doc:"The prefix to use for the cache volume key." default:"gale"`
}

func NewInternalServices() *InternalServices {
	return &InternalServices{
		Artifact: base().WithEntrypoint([]string{"/usr/local/bin/artifact-service"}).
			WithExposedPort(8080).
			WithEnvVariable("PORT", "8080"),
		ArtifactCache: base().WithEntrypoint([]string{"/usr/local/bin/artifactcache-service"}).
			WithExposedPort(8080).
			WithEnvVariable("PORT", "8080"),
	}
}

// BindServices binds the internal services required for running actions
func (is *InternalServices) BindServices(ctx context.Context, c *Container, opts InternalServiceOpts) (container *Container, err error) {
	container = c

	container, err = is.BindDockerService(ctx, container, dag.CacheVolume(fmt.Sprintf("%s-docker", opts.CacheVolumeKeyPrefix)))
	if err != nil {
		return nil, err
	}

	container, err = is.BindArtifactService(ctx, container, dag.CacheVolume(fmt.Sprintf("%s-artifacts", opts.CacheVolumeKeyPrefix)))
	if err != nil {
		return nil, err
	}

	container, err = is.BindArtifactCacheService(ctx, container, dag.CacheVolume(fmt.Sprintf("%s-artifactcache", opts.CacheVolumeKeyPrefix)))
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
	service := is.Artifact.WithMountedCache("/artifacts", volume).WithEnvVariable("ARTIFACTS_DIR", "/artifacts").AsService()

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
	service := is.ArtifactCache.WithMountedCache("/cache", volume).WithEnvVariable("CACHE_DIR", "/cache").AsService()

	endpoint, err := service.Endpoint(ctx, ServiceEndpointOpts{Scheme: "http"})
	if err != nil {
		return nil, err
	}

	// convert the endpoint to expected format for the actions runtime
	endpoint = endpoint + "/"

	return container.WithServiceBinding("artifactcache", service).WithEnvVariable("ACTIONS_CACHE_URL", endpoint), nil
}
