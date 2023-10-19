# Dagger KinD Module

Manage KinD clusters using Dagger.

## Examples

### Create a KinD cluster and Export Kubeconfig

The following query will create a KinD cluster named `my-cluster` and export its kubeconfig to `./config` file:

```shell
dagger query -m github.com/aweris/daggerverse/kind --progress=plain <<EOF
{
    kind {
        cluster (name: "my-cluster") {
            create { 
              kubeconfig (internal: false) { export(path: "./config") }
            }
        }
    }
}
EOF
```

### Run KinD Cli directly to create a cluster and export kubeconfig 

The following query will execute `kind` cli directly with given arguments and calls `stdout` function from the returned
container to get the output:

```shell
dagger query -m github.com/aweris/daggerverse/kind --progress=plain <<EOF
{
    kind {
            exec (args: ["get", "clusters"]) {
                 stdout
            }
    }
}
EOF
```

## Functions

### Kind

#### exec

Execute `kind` cli directly with given arguments and return the container that runs the command.

```graphql
exec(args: [String!]!): Container!
```

Example:

```graphql
{
    kind {
            exec (args: ["get", "clusters"]) {
                 # container object
            }
    }
}
```

#### cluster

Returns a Cluster object that can be used to manage KinD clusters.

```graphql
cluster(name: String!): Cluster!
```

Example:

```graphql
{
    kind {
        cluster (name: "my-cluster") {
            # cluster object
        }
    }
}
```

### Cluster

#### create

Create a KinD cluster with given name. If a cluster with the same name already exists, it won't do anything.

```graphql
create(): Cluster!
```

Example:

```graphql
{
    kind {
        cluster (name: "my-cluster") {
            create {
                # cluster object
            }
        }
    }
}
```

#### kubeconfig

Returns a Kubeconfig object that can be used to manage the kubeconfig of the cluster.

```graphql
kubeconfig(
    internal: Boolean! # if set to true, it will return the internal kubeconfig of the cluster. Default is false. 
): File!
```

When `internal` is set to `true`, it will return the internal kubeconfig of the cluster. It is useful when you want to
use the cluster from inside the dagger containers. When `internal` is set to `false`, it will return the kubeconfig
that can be used from outside of the dagger containers. 

Example:

```graphql
{
    kind {
        cluster (name: "my-cluster") {
            create {
                kubeconfig (internal: false) {
                    # kubeconfig File
                }
            }
        }
    }
}
```

#### logs

Returns a directory contains the logs of the cluster.

```graphql
logs(): Directory!
```

Example:

```graphql
{
    kind {
        cluster (name: "my-cluster") {
            create {
                logs {
                    # directory object
                }
            }
        }
    }
}
```

#### delete

Delete the cluster.

```graphql
delete(): Cluster!
```

Example:

```graphql
{
    kind {
        cluster (name: "my-cluster") {
            create {
                delete {
                    # cluster object
                }
            }
        }
    }
}
```


## Limitations

The Kind module makes DooD (Docker out of ~~Docker~~Dagger) to work. It means that the Docker daemon running on the host
machine is used to run the KinD cluster. It is not fully isolated from the host machine.