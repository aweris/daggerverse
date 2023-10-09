package main

import (
	"context"
)

// serviceArtifacts returns the artifacts service endpoint and the service.
func (g *Gale) serviceArtifacts(ctx context.Context) (endpoint string, service *Service, err error) {
	data := dag.CacheVolume("zenith-gale-artifacts")

	service = g.container().
		WithEntrypoint([]string{"/usr/local/bin/artifact-service"}).
		WithMountedCache("/data", data).WithEnvVariable("ARTIFACTS_DIR", "/data").
		WithExposedPort(8080).WithEnvVariable("PORT", "8080").
		Service()

	endpoint, err = service.Endpoint(ctx, ServiceEndpointOpts{Scheme: "http"})
	if err != nil {
		return "", nil, err
	}

	// convert the endpoint to a expected format for the actions runtime
	endpoint = endpoint + "/"

	return endpoint, service, nil
}

// serviceArtifactCache returns the artifact cache service endpoint and the service.
func (g *Gale) serviceArtifactCache(ctx context.Context) (endpoint string, service *Service, err error) {
	data := dag.CacheVolume("zenith-gale-cache")

	service = g.container().
		WithEntrypoint([]string{"/usr/local/bin/artifactcache-service"}).
		WithMountedCache("/data", data).WithEnvVariable("CACHE_DIR", "/data").
		WithExposedPort(8080).WithEnvVariable("PORT", "8080").
		Service()

	endpoint, err = service.Endpoint(ctx, ServiceEndpointOpts{Scheme: "http"})
	if err != nil {
		return "", nil, err
	}

	// convert the endpoint to a expected format for the actions runtime
	endpoint = endpoint + "/"

	return endpoint, service, nil
}
