# Dagger Kubectl Module

Run kubectl commands using Dagger.

## Examples

### Run kubectl command directly

Run the following command to create a KinD cluster named `my-cluster` and get its kubeconfig:
 
```bash
config=$(echo '{
    kind {
        cluster (name: "my-cluster") {
            create { 
              kubeconfig (internal: true) { id }
            }
        }
    }
}' | dagger query -m github.com/aweris/daggerverse/kind --progress=plain)

fileID=$(echo $config | jq -r '.kind.cluster.create.kubeconfig.id')

echo '{
    kubectl {
        cli(config:"'$fileID'") {
          exec (args: ["get", "pods", "-A", "-o", "json"])
        }
    }
}' | dagger query -m github.com/aweris/daggerverse/kubectl --progress=plain
```

The command above will:

- Create a KinD cluster named `my-cluster` (if it does not exist)
- Get internal kubeconfig file for the cluster and return its ID
- Create a kubectl client using the kubeconfig file ID
- Execute `kubectl get pods -A -o json` command and return the result

### Get a kubectl container and run commands

```bash
config=$(echo '{
    kind {
        cluster (name: "my-cluster") {
            create { 
              kubeconfig (internal: true) { id }
            }
        }
    }
}' | dagger query -m github.com/aweris/daggerverse/kind --progress=plain)

fileID=$(echo $config | jq -r '.kind.cluster.create.kubeconfig.id')

namespace="my-namespace-$(date +%s)"

echo '{
    kubectl {
        cli(config:"'$fileID'") {
          container {
            withExec(args: ["create", "namespace", "'$namespace'"]) {
              withExec(args: ["get", "namespaces"]) {
                stdout
              }
            }
          }
        }
    }
}' | dagger query -m github.com/aweris/daggerverse/kubectl --progress=plain
```

The command above will:

- Create a KinD cluster named `my-cluster` (if it does not exist)
- Get internal kubeconfig file for the cluster and return its ID
- Create a kubectl client using the kubeconfig file ID
- Get a kubectl container
- Execute `kubectl create namespace <namespace>` and `kubectl get namespaces` commands in the container 
- Return the stdout of the last command