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

## Releases

Pre-built binaries for Linux, macOS, and Windows are published on the
[GitHub Releases](../../releases) page. Each release contains an archive per
platform/architecture plus a `checksums.txt` with SHA-256 sums.

Print the version of an installed binary with:

```sh
kubegrep version
```

### Versioning

Releases follow [Semantic Versioning](https://semver.org/) (`vMAJOR.MINOR.PATCH`):

- **MAJOR** — incompatible/breaking changes.
- **MINOR** — backwards-compatible new functionality.
- **PATCH** — backwards-compatible bug fixes.

Pre-releases use a suffix, e.g. `v1.2.0-rc.1`.

### Cutting a new release

Releases are fully automated by the
[`Release` workflow](.github/workflows/release.yml), which is triggered by
pushing a semver tag. To publish a new version:

1. Make sure `main` is green (the `CI` workflow builds and tests every push).
2. Choose the next version number according to semver, e.g. `v1.2.3`.
3. Create and push an annotated tag:

   ```sh
   git tag -a v1.2.3 -m "kubegrep v1.2.3"
   git push origin v1.2.3
   ```

4. The `Release` workflow then:
   - cross-compiles the binary for each supported platform,
   - packages each one into a `.tar.gz` (or `.zip` for Windows) including
     `README.md` and `LICENSE`,
   - generates `checksums.txt`,
   - creates a GitHub Release with auto-generated release notes and attaches
     all the artifacts.

The version reported by `kubegrep version` is injected from the tag name at
build time, so released binaries report their exact version.
