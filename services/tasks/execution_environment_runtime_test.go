package tasks

import (
	"context"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
)

func TestExecutionEnvironmentCommand(t *testing.T) {
	env := db.ExecutionEnvironment{
		Name: "ansible",
		Image: "registry.example.test/ansible@sha256:" + strings.Repeat("a", 64),
		PullPolicy: db.ExecutionEnvironmentPullIfNotPresent,
	}
	job := ExecutionEnvironmentJob{Environment: env, WorkingDirectory: "/tmp/ee-work", Arguments: []string{"ansible-playbook", "site.yml"}}
	cmd, err := (ExecutionEnvironmentRuntime{Binary: "podman"}).command(context.Background(), job)
	if err != nil { t.Fatal(err) }
	joined := strings.Join(cmd.Args, " ")
	for _, want := range []string{"--pull=missing", "--network=none", "--cap-drop=ALL", "--read-only", "--user=65534:65534", "--memory=1g", "--cpus=1", env.Image} {
		if !strings.Contains(joined, want) { t.Errorf("missing runtime restriction %q: %s", want, joined) }
	}
	for _, binary := range []string{"sh", "bash", "", "nerdctl"} {
		if _, err := (ExecutionEnvironmentRuntime{Binary: binary}).command(context.Background(), job); err == nil {
			t.Errorf("unexpected acceptance of %q", binary)
		}
	}
	job.WorkingDirectory = "../escape"
	if _, err := (ExecutionEnvironmentRuntime{Binary: "docker"}).command(context.Background(), job); err == nil {
		t.Fatal("relative workspace accepted")
	}
}
