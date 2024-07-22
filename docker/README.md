# Dagger Docker Module

![dagger-min-version](https://img.shields.io/badge/dagger%20version-v0.12.1-green)

Docker module provides DinD using Dagger.

**Note:** This module is experimental and please use it with caution.

## Prerequisites

- Module requires Dagger CLI version `v0.12.1` or higher.

## Limitations

This module requires to run Docker service with `InsecureRootCapabilities` enabled. This means that container started
with `--privileged` flag. This is a security risk and should be used with caution.
