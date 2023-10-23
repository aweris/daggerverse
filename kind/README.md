# Dagger KinD Module

Easily manage KinD clusters through Dagger.

## Before you start

Set `DAGGER_MODULE` to environment variable to avoid using `-m github.com/aweris/daggerverse/kind` in every command.

```shell
export DAGGER_MODULE=github.com/aweris/daggerverse/kind
```

## Commands

### Create a KinD cluster

```shell
dagger call cluster --name my-cluster create
```

- If no name is given, it defaults to kind.
- if cluster already exists, it won't be created again. It is safe to call this command multiple times.

### Export Kubeconfig

Save the kubeconfig file to ./config:

For connecting from host: 

```shell
dagger download --name my-cluster kubeconfig --export-path ./config
```
For connecting from inside the cluster:

```shell
dagger download cluster --name my-cluster kubeconfig --internal --export-path ./config
```

### Download Cluster Logs 

Save cluster logs to `./logs` directory:

```shell
dagger download cluster --name my-cluster logs --export-path ./logs
```

### Delete a Cluster

```shell
dagger call cluster --name my-cluster delete
```

### Command Shells

For KinD cli:

```shell
dagger shell cluster --name my-cluster cli
```

For K9S:

```shell
dagger shell -m github.com/aweris/daggerverse/kind cluster --name my-cluster k-9-s
```

## Limitations

This module requires to access the host's Docker daemon. This means that KinD clusters not completely isolated from the
host. 