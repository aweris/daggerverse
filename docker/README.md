# Dagger Docker Module

![dagger-min-version](https://img.shields.io/badge/dagger%20version-v0.9.1-green)

Docker module provides DinD using Dagger.

**Note:** This module is experimental and please use it with caution.

## Prerequisites

- Module requires Dagger CLI version `v0.9.1` or higher.

## Before you start

Set `DAGGER_MODULE` to environment variable to avoid using `-m github.com/aweris/daggerverse/docker` in every command.

```shell
export DAGGER_MODULE=github.com/aweris/daggerverse/docker
```

## Commands

### Start DinD Service

```shell
dagger up dind  -p 2375:2375
```

then set `DOCKER_HOST` to `tcp://localhost:2375` to use Docker CLI. 

## Limitations

This module requires to run Docker service with `InsecureRootCapabilities` enabled. This means that container started 
with `--privileged` flag. This is a security risk and should be used with caution.