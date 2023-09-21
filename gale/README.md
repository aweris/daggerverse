# gale

Dagger module for executing Github Actions workflows using `gale` CLI.

## Requirements

You need to run a [Dagger with module support](https://github.com/shykes/dagger/tree/zenith-functions/zenith#project-zenith)

## Examples

First set `GITHUB_TOKEN` environment variable to your Github token.

```bash
export GITHUB_TOKEN=<super-secret-token>
```
### List workflows

```bash
dagger query --progress=plain list < <(sed "s/\$GITHUB_TOKEN/$GITHUB_TOKEN/g" examples.gql)
```

### Run workflow

```bash
dagger query --progress=plain run < <(sed "s/\$GITHUB_TOKEN/$GITHUB_TOKEN/g" examples.gql)
```