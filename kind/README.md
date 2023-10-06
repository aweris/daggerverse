# Dagger KinD Module

Manage KinD clusters using Dagger.

## Quick Start: 

For a quick start, run the following command to create a KinD cluster named `my-cluster` and export its kubeconfig to
`./config` file:
 
```bash
echo '{
    kind {
        cluster (name: "my-cluster") {
            create,
            kubeconfig (internal: false) { export(path: "./config") }
        }
    }
}' | dagger query -m github.com/aweris/daggerverse/kind --progress=plain
```

or run individual kind cli commands:

```bash
echo '{
    kind { 
        exec (args: ["create", "cluster", "--name", "my-cluster"]) { stdout }
        
        exec (args: ["export", "kubeconfig", "--name", "my-cluster"]) {
            file(path: "/root/.kube/config") {  export(path: "./kubeconfig") }
        }
    }
}' | dagger query -m github.com/aweris/daggerverse/kind --progress=plain
```

The above command will run `kind` command given in the `args` field and return dagger container executed that command. 
With `stdout` field, we are asking dagger to return the stdout of the command.

## Limitations

This module requires to access the host's Docker daemon. This means that KinD clusters not completely isolated from the
host. 