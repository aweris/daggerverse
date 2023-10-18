# Dagger Gale Module

Dagger module for executing Github Actions workflows using `gale`.

## Requirements

The module uses [services v2](https://github.com/dagger/dagger/pull/5557) to create internal services and runs modules
using [project-zenith](https://github.com/shykes/dagger/tree/zenith-functions/zenith#project-zenith). This means that
until `services v2` is merged and released, you need to build Dagger from [services-v2-c2h-h2c](https://github.com/vito/dagger/tree/services-v2-c2h-h2c) 
branch.

## How to use Gale

### Building Dagger from source

- Build Dagger from [services-v2-c2h-h2c](https://github.com/vito/dagger/tree/services-v2-c2h-h2c) branch.

```shell
git clone git@github.com:vito/dagger.git --branch services-v2-c2h-h2c

cd dagger

./hack/dev
```

- Create a `.envrc` in the root of the directory you want to run Gale from.

```shell
DAGGER_SRC_ROOT=<path to dagger source root>
export _EXPERIMENTAL_DAGGER_RUNNER_HOST=docker-container://dagger-engine.dev
export PATH=$DAGGER_SRC_ROOT/bin:$PATH
```

- Run `direnv allow` to load the environment variables.

```shell
direnv allow
```

Now you can run `dagger` from the root of the directory with `Services v2` enabled.

### Running Gale

- Gale requires a Github token to be able to access Github API. First step you need to set `GITHUB_TOKEN` environment

```shell
export GITHUB_TOKEN=<your github token>
```

or use `gh cli` to set it:

```shell
export GITHUB_TOKEN=$(gh auth token)
```

#### List workflows

List workflows for a repository `aweris/gale`:

```shell
dagger query -m github.com/aweris/daggerverse/gale <<EOF
{
    gale {
        setToken(token: "$GITHUB_TOKEN") {
            repo(repo: "aweris/gale") {
                workflows {
                  list
                }
            }
        }
    }
}
EOF
```

#### Run a workflow

Run a workflow `build` workflows `lint` job for the `kubernetes/minikube` repo :

```shell
dagger query -m github.com/aweris/daggerverse/gale <<EOF
{
    gale {
        setToken(token: "$GITHUB_TOKEN") {
            repo(repo: "kubernetes/minikube") {
                workflows {
                    run(workflow: "build", job: "lint") {
                       result
                    }
                }
            }
        }
    }
}
EOF
```

## Functions

### Gale

#### SetToken

Sets the Github token used for authentication. This method needs to be called before any other method to be able to
access Github API.

```graphql
setToken(token: String!): Gale
```

Example GraphQL query:

```graphql
{
    gale {
        setToken(token: "$GITHUB_TOKEN") {
            # Gale with auth token set
        }
    }
}
```

#### Repo

Returns a repository object for the given repository name.

```graphql
repo(
    repo:   String    # Name of the repository. name can be in the following formats: OWNER/REPO, HOST/OWNER/REPO, and a full URL.
    source: Directory # Directory containing the repository
    tag:    String    # Tag to checkout
    branch: String    # Branch to checkout
    commit: String    # Commit to checkout
): Repo
```

All parameters are optional. Only requirement is that `repo` or `source` must be set. If both are set, `repo` takes
precedence.

| Repo | Source | Tag/Branch/Commit | Result                            |
|------|--------|-------------------|-----------------------------------|
| n/a  | n/a    | n/a               | Error                             |
| set  | n/a    | n/a               | Repo with default branch          |
| n/a  | set    | n/a               | Source directory as it is         |
| set  | set    | n/a               | Repo with default branch          |
| set  | set    | set               | Repo with given tag/branch/commit |

**Note:** If `tag`, `branch` and `commit` are set, predicate is `tag` > `branch` > `commit`.
 
Example GraphQL query:

```graphql
{
    gale {
        setToken(token: "$GITHUB_TOKEN") {
            repo(repo: "kubernetes/minikube") {
                # Repo object
            }
        }
    }
}
```

### Repo

#### Workflows

Returns a workflows object for the repository.

```graphql
workflows(
  workflowsDir: String # Directory containing the workflows. Default is `.github/workflows`
): Workflows
```

Example GraphQL query:

```graphql
{
    gale {
        setToken(token: "$GITHUB_TOKEN") {
            repo(repo: "kubernetes/minikube") {
                workflows(workflowsDir: "some/other/dir") {
                    # Workflows object
                }
            }
        }
    }
}
```

### Workflows

#### List

Prints a list of workflows and their jobs in the workflows directory.

```graphql
list
```

Example GraphQL query:

```graphql
{
    gale {
        setToken(token: "$GITHUB_TOKEN") {
            repo(repo: "kubernetes/minikube") {
                workflows {
                    list
                }
            }
        }
    }
}
```

#### Run

Creates a workflow run object for the given options and runs the workflow.

```graphql
run(
    workflow:     String! # Name of the workflow to run
    job:          String  # Name of the job to run, if not set, all jobs in the workflow will be run
    eventName:    String  # Name of the event to trigger the workflow. Default is `push`
    eventFile:    File    # File containing the event payload.
    runnerImage:  String  # Docker image to use for the runner. Default is `ghcr.io/catthehacker/ubuntu:act-latest`
    debug:        Boolean # Enable debug mode. Default is `false`
    dind:         Boolean # Start docker-in-dagger service to isolate docker daemon. Default is `false`
    dockerSocket: String  # Path to docker socket. Only used if `dind` is not `true`. Default is `/var/run/docker.sock`
): WorkflowRun
```

Example GraphQL query:

```graphql
{
    gale {
        setToken(token: "$GITHUB_TOKEN") {
            repo(repo: "kubernetes/minikube") {
                workflows {
                    run(workflow: "build", job: "lint") {
                       # WorkflowRun object
                    }
                }
            }
        }
    }
}
```

### WorkflowRun

#### Result

Executes the workflow and returns the result summary as a string.

```graphql
result: String
```

Example GraphQL query:

```graphql
{
    gale {
        setToken(token: "$GITHUB_TOKEN") {
            repo(repo: "kubernetes/minikube") {
                workflows {
                    run(workflow: "build", job: "lint") {
                       result
                    }
                }
            }
        }
    }
}
```

Example result output:
```shell
...
<Workflow Logs>
...

410: query
410: [95.5s] {
410: [95.5s]     "gale": {
410: [95.5s]         "setToken": {
410: [95.5s]             "repo": {
410: [95.5s]                 "workflows": {
410: [95.5s]                     "run": {
410: [95.5s]                         "result": "Workflow build completed with conclusion failure in 1m28.598518623s"
410: [95.5s]                     }
410: [95.5s]                 }
410: [95.5s]             }
410: [95.5s]         }
410: [95.5s]     }
410: [95.5s] }
410: query DONE
```

#### Sync

Executes the workflow and returns the runner container itself. This method is useful if you want to access the runner 
container after the workflow is completed.

```graphql
sync: Container
```

Example GraphQL query:

```graphql
{
    gale {
        setToken(token: "$GITHUB_TOKEN") {
            repo(repo: "kubernetes/minikube") {
                workflows {
                    run(workflow: "build", job: "lint") {
                       sync {
                           # Container object
                       }
                    }
                }
            }
        }
    }
}
```

#### Directory

Returns the directories used by the workflow. If no options are set, it returns `runs` directory containing the workflow
related files.

```graphql
directory(
    includeRepo:      Boolean # Include the repository source in the exported directory. Default is `false`
    includeMetadata:  Boolean # Include the workflow run metadata in the exported directory. Default is `false`
    includeSecrets:   Boolean # Include the secrets in the exported directory. Default is `false`
    includeEvent:     Boolean # Include the event file in the exported directory. Default is `false`
    includeActions:   Boolean # Include the custom action repos used in runner in the exported directory. Default is `false`
    includeArtifacts: Boolean # Include the artifacts in the exported directory. Default is `false`
): Directory
```

Example GraphQL query:

```graphql
{
    gale {
        setToken(token: "$GITHUB_TOKEN") {
            repo(repo: "kubernetes/minikube") {
                workflows {
                    run(workflow: "build", job: "lint") {
                       directory {
                           # Directory object
                       }
                    }
                }
            }
        }
    }
}
```