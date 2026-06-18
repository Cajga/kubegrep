package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// resourceMeta is used to extract the kind and metadata.name from a
// kubernetes resource without losing the rest of the document.
type resourceMeta struct {
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "kubegrep:", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) > 0 && args[0] == "completion" {
		return completion(args[1:], stdout)
	}

	fs := flag.NewFlagSet("kubegrep", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var kind string
	fs.StringVar(&kind, "kind", "", "filter resources by kind (case insensitive)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: kubegrep [--kind KIND] [NAME_PATTERN] [FILE]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Grep-like filtering of kubernetes resources from YAML manifests.")
		fmt.Fprintln(os.Stderr, "Reads from FILE or standard input when FILE is omitted.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Run 'kubegrep completion bash' to generate a bash completion script.")
		fmt.Fprintln(os.Stderr, "")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	namePattern, filePath, err := resolveArgs(fs.Args(), fileExists)
	if err != nil {
		return err
	}

	var input io.Reader = stdin
	if filePath != "" {
		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close()
		input = f
	}

	return grep(input, stdout, kind, namePattern)
}

// completion writes a shell completion script to out for the requested shell.
// Currently only bash is supported, mirroring the "completion SHELL" command
// found in tools like kubectl and kustomize.
func completion(args []string, out io.Writer) error {
	if len(args) != 1 {
		return errors.New("usage: kubegrep completion bash")
	}
	switch args[0] {
	case "bash":
		fmt.Fprint(out, bashCompletion)
		return nil
	default:
		return fmt.Errorf("unsupported shell %q (only \"bash\" is supported)", args[0])
	}
}

// bashCompletion is the bash completion script for kubegrep.
const bashCompletion = `# bash completion for kubegrep
# To load completions in your current shell session:
#   source <(kubegrep completion bash)
# To load completions for every new session, on Linux:
#   kubegrep completion bash > /etc/bash_completion.d/kubegrep
# on macOS:
#   kubegrep completion bash > $(brew --prefix)/etc/bash_completion.d/kubegrep

_kubegrep() {
    local cur prev words cword
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    local flags="--kind -h --help"
    local commands="completion"

    # Complete the argument to the completion command.
    if [[ "${COMP_WORDS[1]}" == "completion" && ${COMP_CWORD} -eq 2 ]]; then
        COMPREPLY=( $(compgen -W "bash" -- "${cur}") )
        return 0
    fi

    case "${prev}" in
        --kind)
            # Resource kinds are free-form; offer a few common ones.
            COMPREPLY=( $(compgen -W "ConfigMap Deployment Service Secret Pod Ingress StatefulSet DaemonSet Job CronJob" -- "${cur}") )
            return 0
            ;;
    esac

    if [[ "${cur}" == -* ]]; then
        COMPREPLY=( $(compgen -W "${flags}" -- "${cur}") )
        return 0
    fi

    if [[ ${COMP_CWORD} -eq 1 ]]; then
        COMPREPLY=( $(compgen -W "${commands}" -- "${cur}") )
    fi

    # Fall back to file completion (for the optional FILE argument).
    COMPREPLY+=( $(compgen -f -- "${cur}") )
    return 0
}

complete -F _kubegrep kubegrep
`

// fileExists reports whether path points to an existing, regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// resolveArgs interprets the positional arguments as an optional NAME_PATTERN
// followed by an optional FILE.
//
// Because NAME_PATTERN is optional, a single positional argument is ambiguous.
// To avoid reading from stdin (and appearing to hang) when the user actually
// passed a manifest path, a lone positional argument that refers to an existing
// file is treated as FILE rather than NAME_PATTERN.
func resolveArgs(rest []string, exists func(string) bool) (namePattern, filePath string, err error) {
	switch len(rest) {
	case 0:
		return "", "", nil
	case 1:
		if exists(rest[0]) {
			return "", rest[0], nil
		}
		return rest[0], "", nil
	case 2:
		return rest[0], rest[1], nil
	default:
		return "", "", errors.New("too many arguments")
	}
}

func grep(input io.Reader, out io.Writer, kind, namePattern string) error {
	dec := yaml.NewDecoder(input)

	for {
		var node yaml.Node
		err := dec.Decode(&node)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		// Skip empty documents (e.g. trailing "---").
		if node.Kind == 0 || (node.Kind == yaml.DocumentNode && len(node.Content) == 0) {
			continue
		}

		var meta resourceMeta
		if err := node.Decode(&meta); err != nil {
			// Not a mapping we can understand; skip it.
			continue
		}

		if !matches(meta, kind, namePattern) {
			continue
		}

		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		if err := enc.Encode(&node); err != nil {
			return err
		}
		if err := enc.Close(); err != nil {
			return err
		}

		fmt.Fprint(out, "---\n")
		if _, err := out.Write(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

func matches(meta resourceMeta, kind, namePattern string) bool {
	if kind != "" && !strings.EqualFold(meta.Kind, kind) {
		return false
	}
	if namePattern != "" && !strings.Contains(
		strings.ToLower(meta.Metadata.Name),
		strings.ToLower(namePattern),
	) {
		return false
	}
	return true
}
