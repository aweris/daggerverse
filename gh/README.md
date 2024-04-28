# Dagger GH module

![dagger-min-version](https://img.shields.io/badge/dagger%20version-v0.11.2-green)

Dagger module for GitHub CLI.

**Note:** This module is experimental and please use it with caution.

## Prerequisites

- KinD module requires Dagger CLI version `v0.11.2` or higher.

## Before you start

Set `DAGGER_MODULE` to environment variable to avoid using `-m github.com/aweris/daggerverse/gh` in every command.

```shell
export DAGGER_MODULE=github.com/aweris/daggerverse/gh
```

## Commands

### Get Github CLI Binary

```shell
dagger call get --version vx.y.z --output ./gh
```

### Running Github CLI

```shell
dagger call run --version vx.y.z --token env:GITHUB_TOKEN --cmd "search repos --followers 1 --json url"
```

## Flags

- `--version`: Version of the GitHub CLI binary to download. Defaults to `v2.37.0`.
- `--token`: GitHub token to use. It requires to run any `gh` command. 