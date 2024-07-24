# Dagger kinD module

![dagger-min-version](https://img.shields.io/badge/dagger%20version-v0.12.1-green)

Dagger module for running [KinD](https://kind.sigs.k8s.io/) clusters locally and in CI/CD pipelines.

[Daggerverse](https://daggerverse.dev/mod/github.com/aweris/daggerverse/kind)

> [!WARNING]
> This module uses the Docker socket on the host due to KinD cluster requirements. It's not fully containerized, which may have security implications. Use caution, especially in production.
