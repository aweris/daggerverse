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

## Limitations

This module requires to access the host's Docker daemon. This means that KinD clusters not completely isolated from the
host. 