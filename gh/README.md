# Dagger GH module

![dagger-min-version](https://img.shields.io/badge/dagger%20version-v0.9.2-green)

Dagger module for GitHub CLI.

**Note:** This module is experimental and please use it with caution.

## Prerequisites

- KinD module requires Dagger CLI version `v0.9.2` or higher.

## Before you start

Set `DAGGER_MODULE` to environment variable to avoid using `-m github.com/aweris/daggerverse/gh` in every command.

```shell
export DAGGER_MODULE=github.com/aweris/daggerverse/gh
```

## Commands

### Get Github CLI Binary

```shell
dagger download gh --version vx.y.z --export-path ./gh
```

### Running Github CLI

```shell
dagger run gh --version vx.y.z --token <token> --cmd "search repos --followers 1 --json url"
```

## Flags

- `--version`: Version of the GitHub CLI binary to download. Defaults to `v2.37.0`.
- `--token`: GitHub token to use. It requires to run any `gh` command. 