# kubegrep

`grep` for Kubernetes resources inside YAML manifests.

It reads a (possibly multi-document) YAML manifest from a file or standard
input, filters the resources, and prints each matching resource as YAML.

## Build

```sh
go build -o kubegrep .
```

## Usage

```
kubegrep [--kind KIND] [NAME_PATTERN] [FILE]
```

- `--kind KIND` — keep only resources whose `kind` equals `KIND`
  (case insensitive).
- `NAME_PATTERN` — optional positional pattern matched as a case-insensitive
  substring against `metadata.name`.
- `FILE` — optional manifest path. When omitted, the manifest is read from
  standard input.

When a single positional argument is given without a NAME_PATTERN and it refers
to an existing file, it is treated as `FILE` (so `kubegrep --kind deployment
manifests.yaml` reads the file instead of waiting on standard input).

If both a kind and a name pattern are given, a resource must satisfy both.
Matching resources are printed in full as YAML, each preceded by a `---`
document separator.

## Examples

```sh
# All Deployments in a file
kubegrep --kind deployment manifests.yaml

# Any resource whose name contains "myapp" (read from stdin)
kubectl get all -o yaml | kubegrep myapp

# Services named like "other"
kubegrep --kind Service other manifests.yaml
```

## Shell completion

Generate a bash completion script the same way `kubectl` and `kustomize` do:

```sh
# Load completions in the current shell session
source <(kubegrep completion bash)

# Install completions for every new session (Linux)
kubegrep completion bash > /etc/bash_completion.d/kubegrep

# Install completions for every new session (macOS, Homebrew)
kubegrep completion bash > "$(brew --prefix)/etc/bash_completion.d/kubegrep"
```
