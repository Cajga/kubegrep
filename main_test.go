package main

import (
	"strings"
	"testing"
)

const manifests = `apiVersion: v1
kind: ConfigMap
metadata:
  name: MyApp-Config
data:
  key: value
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-deploy
spec:
  replicas: 2
---
apiVersion: v1
kind: Service
metadata:
  name: other-svc
`

func runGrep(t *testing.T, kind, name string) string {
	t.Helper()
	var out strings.Builder
	if err := grep(strings.NewReader(manifests), &out, kind, name); err != nil {
		t.Fatalf("grep failed: %v", err)
	}
	return out.String()
}

func TestKindCaseInsensitive(t *testing.T) {
	got := runGrep(t, "deployment", "")
	if !strings.Contains(got, "myapp-deploy") {
		t.Errorf("expected deployment, got: %s", got)
	}
	if strings.Contains(got, "other-svc") || strings.Contains(got, "MyApp-Config") {
		t.Errorf("unexpected extra resources: %s", got)
	}
}

func TestNameCaseInsensitive(t *testing.T) {
	got := runGrep(t, "", "myapp")
	if !strings.Contains(got, "MyApp-Config") || !strings.Contains(got, "myapp-deploy") {
		t.Errorf("expected both myapp resources, got: %s", got)
	}
	if strings.Contains(got, "other-svc") {
		t.Errorf("unexpected service in output: %s", got)
	}
}

func TestKindAndName(t *testing.T) {
	got := runGrep(t, "Service", "other")
	if !strings.Contains(got, "other-svc") {
		t.Errorf("expected service, got: %s", got)
	}
	if strings.Contains(got, "myapp") {
		t.Errorf("unexpected match: %s", got)
	}
}

func TestNoMatch(t *testing.T) {
	got := runGrep(t, "Ingress", "")
	if strings.TrimSpace(got) != "" {
		t.Errorf("expected empty output, got: %s", got)
	}
}

func TestResolveArgsSinglePositionalIsFileWhenItExists(t *testing.T) {
	// "manifests.yaml" refers to an existing file, so without a NAME_PATTERN
	// it must be treated as FILE rather than being mistaken for the pattern
	// (which would leave the manifest unread from stdin and appear to hang).
	exists := func(p string) bool { return p == "manifests.yaml" }

	name, file, err := resolveArgs([]string{"manifests.yaml"}, exists)
	if err != nil {
		t.Fatalf("resolveArgs failed: %v", err)
	}
	if name != "" {
		t.Errorf("expected empty NAME_PATTERN, got %q", name)
	}
	if file != "manifests.yaml" {
		t.Errorf("expected FILE %q, got %q", "manifests.yaml", file)
	}
}

func TestResolveArgsSinglePositionalIsPatternWhenNotAFile(t *testing.T) {
	exists := func(string) bool { return false }

	name, file, err := resolveArgs([]string{"myapp"}, exists)
	if err != nil {
		t.Fatalf("resolveArgs failed: %v", err)
	}
	if name != "myapp" {
		t.Errorf("expected NAME_PATTERN %q, got %q", "myapp", name)
	}
	if file != "" {
		t.Errorf("expected empty FILE, got %q", file)
	}
}

func TestResolveArgsNameAndFile(t *testing.T) {
	name, file, err := resolveArgs([]string{"myapp", "manifests.yaml"}, func(string) bool { return true })
	if err != nil {
		t.Fatalf("resolveArgs failed: %v", err)
	}
	if name != "myapp" || file != "manifests.yaml" {
		t.Errorf("expected (%q, %q), got (%q, %q)", "myapp", "manifests.yaml", name, file)
	}
}

func TestCompletionBash(t *testing.T) {
	var out strings.Builder
	if err := run([]string{"completion", "bash"}, strings.NewReader(""), &out); err != nil {
		t.Fatalf("completion bash failed: %v", err)
	}
	got := out.String()
	for _, want := range []string{"_kubegrep()", "complete -F _kubegrep kubegrep", "--kind"} {
		if !strings.Contains(got, want) {
			t.Errorf("completion output missing %q, got: %s", want, got)
		}
	}
}

func TestCompletionUnsupportedShell(t *testing.T) {
	var out strings.Builder
	if err := run([]string{"completion", "zsh"}, strings.NewReader(""), &out); err == nil {
		t.Errorf("expected error for unsupported shell")
	}
}

func TestCompletionMissingShell(t *testing.T) {
	var out strings.Builder
	if err := run([]string{"completion"}, strings.NewReader(""), &out); err == nil {
		t.Errorf("expected error when shell is omitted")
	}
}

func TestResolveArgsTooMany(t *testing.T) {
	if _, _, err := resolveArgs([]string{"a", "b", "c"}, func(string) bool { return false }); err == nil {
		t.Errorf("expected error for too many arguments")
	}
}
