package db

import (
	"strings"
	"testing"
)

func TestExecutionEnvironmentValidate(t *testing.T) {
	valid := ExecutionEnvironment{
		Name: "ansible-ee",
		Image: "registry.example.test/automation/ansible@sha256:" + strings.Repeat("a", 64),
		PullPolicy: ExecutionEnvironmentPullIfNotPresent,
	}
	if err := valid.Validate(); err != nil { t.Fatalf("valid pinned environment rejected: %v", err) }
	cases := []struct {
		name string
		edit func(*ExecutionEnvironment)
	}{
		{"mutable tag", func(e *ExecutionEnvironment) { e.Image = "registry.example.test/ansible:latest" }},
		{"short digest", func(e *ExecutionEnvironment) { e.Image = "registry.example.test/ansible@sha256:abcd" }},
		{"uppercase digest", func(e *ExecutionEnvironment) { e.Image = "registry.example.test/ansible@sha256:" + strings.Repeat("A", 64) }},
		{"empty name", func(e *ExecutionEnvironment) { e.Name = "" }},
		{"whitespace name", func(e *ExecutionEnvironment) { e.Name = " ansible" }},
		{"shell whitespace", func(e *ExecutionEnvironment) { e.Image += " --privileged" }},
		{"unsupported pull policy", func(e *ExecutionEnvironment) { e.PullPolicy = "force" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ee := valid
			tc.edit(&ee)
			if err := ee.Validate(); err == nil { t.Fatalf("expected invalid execution environment: %+v", ee) }
		})
	}
}
