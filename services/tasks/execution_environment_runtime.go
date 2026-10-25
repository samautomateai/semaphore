package tasks

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/semaphoreui/semaphore/db"
)

// ExecutionEnvironmentRuntime is an explicit, opt-in OCI runtime adapter.
// It does not enable execution environments in task dispatch by itself.
type ExecutionEnvironmentRuntime struct {
	Binary string // "podman" or "docker"
}

type ExecutionEnvironmentJob struct {
	Environment db.ExecutionEnvironment
	WorkingDirectory string
	Arguments []string
}

func (r ExecutionEnvironmentRuntime) command(ctx context.Context, job ExecutionEnvironmentJob) (*exec.Cmd, error) {
	if err := job.Environment.Validate(); err != nil {
		return nil, err
	}
	if r.Binary != "podman" && r.Binary != "docker" {
		return nil, errors.New("unsupported OCI runtime")
	}
	if job.WorkingDirectory == "" || !strings.HasPrefix(job.WorkingDirectory, "/") || strings.Contains(job.WorkingDirectory, ":") {
		return nil, errors.New("working directory must be an absolute Linux path")
	}
	if len(job.Arguments) == 0 {
		return nil, errors.New("container command must be explicit")
	}
	for _, arg := range job.Arguments {
		if strings.ContainsRune(arg, 0) {
			return nil, errors.New("container command contains NUL")
		}
	}
	args := []string{
		"run", "--rm", "--pull=" + string(job.Environment.PullPolicy),
		"--network=none", "--cap-drop=ALL", "--security-opt=no-new-privileges",
		"--read-only", "--pids-limit=256", "--memory=1g", "--cpus=1",
		"--user=65534:65534", "--workdir=/workspace",
		"--mount", "type=bind,src=" + job.WorkingDirectory + ",dst=/workspace,readonly",
		job.Environment.Image,
	}
	args = append(args, job.Arguments...)
	return exec.CommandContext(ctx, r.Binary, args...), nil
}

// Run invokes the configured runtime with an argument vector, never a shell.
// Callers must provision a dedicated workspace with safe ownership and cleanup.
func (r ExecutionEnvironmentRuntime) Run(ctx context.Context, job ExecutionEnvironmentJob) error {
	cmd, err := r.command(ctx, job)
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("execution environment failed: %w", err)
	}
	return nil
}
