# Dagger Helm Module

![dagger-version](https://img.shields.io/badge/dagger%20version-v0.9.0-green)

Execute Helm commands with Dagger.

**Note:** This module is experimental and please use it with caution.

## Examples

### Run Helm Command Directly

Add a repository and install a Helm chart:

```bash
# Create a KinD cluster and get kubeconfig ID
config=$(dagger query -m github.com/aweris/daggerverse/kind --progress=plain <<EOF
{
  kind {
    cluster (name: "my-cluster") {
      create { kubeconfig (internal: true) { id } }
    }
  }
}
EOF
)

fileID=$(echo "$config" | jq -r '.kind.cluster.create.kubeconfig.id')

# Add repo and install chart
dagger query -m github.com/aweris/daggerverse/helm --progress=plain <<EOF
{
  helm {
    cli(config:"$fileID") {
      exec (args: ["repo", "add", "bitnami", "https://charts.bitnami.com/bitnami"])
    }
  }
}
EOF

releaseName="my-release-$(date +%s)"

dagger query -m github.com/aweris/daggerverse/helm --progress=plain <<EOF
{
  helm {
    cli(config:"$fileID") {
      exec (args: ["install", "$releaseName", "bitnami/nginx", "--namespace", "my-namespace", "--create-namespace"])
    }
  }
}
EOF
```

### Get a Helm Container and Run Commands

Execute multiple commands in a Helm container:

```bash
# Create a KinD cluster and get kubeconfig ID
config=$(dagger query -m github.com/aweris/daggerverse/kind --progress=plain <<EOF
{
  kind {
    cluster (name: "my-cluster") {
      create { kubeconfig (internal: true) { id } }
    }
  }
}
EOF
)

fileID=$(echo "$config" | jq -r '.kind.cluster.create.kubeconfig.id')

# Get a Helm container and run commands
releaseName="my-release-$(date +%s)"

dagger query -m github.com/aweris/daggerverse/helm --progress=plain <<EOF
{
  helm {
    cli(config:"$fileID") {
      container {
        withExec(args: ["repo", "add", "bitnami", "https://charts.bitnami.com/bitnami"]) {
          withExec(args: ["install", "$releaseName", "bitnami/nginx", "--namespace", "my-namespace", "--create-namespace"]) {
            withExec(args: ["list", "-A"]) {
              stdout
            }
          }
        }
      }
    }
  }
}
EOF
```
