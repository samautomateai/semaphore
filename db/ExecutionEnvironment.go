package db

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ExecutionEnvironment defines a pinned, immutable container image for an
// Ansible execution worker. This model deliberately does not imply that
// container execution is already wired into the task scheduler.
type ExecutionEnvironment struct {
	Name string `json:"name"`
	ProjectID int `json:"project_id"`
	RegistryCredentialID *int `json:"registry_credential_id,omitempty"`
	Image string `json:"image"`
	PullPolicy ExecutionEnvironmentPullPolicy `json:"pull_policy"`
}

type ExecutionEnvironmentPullPolicy string

const (
	ExecutionEnvironmentPullAlways ExecutionEnvironmentPullPolicy = "always"
	ExecutionEnvironmentPullIfNotPresent ExecutionEnvironmentPullPolicy = "if_not_present"
	ExecutionEnvironmentPullNever ExecutionEnvironmentPullPolicy = "never"
)

var executionEnvironmentDigest = regexp.MustCompile(`^[a-f0-9]{64}$`)

// Validate requires a digest-pinned OCI image to prevent mutable tags from
// silently changing what a scheduled automation job executes.
func (e ExecutionEnvironment) Validate() error {
	if strings.TrimSpace(e.Name) != e.Name || e.Name == "" || len(e.Name) > 128 {
		return errors.New("execution environment name must be nonempty, canonical and at most 128 bytes")
	}
	if strings.TrimSpace(e.Image) != e.Image || strings.ContainsAny(e.Image, " \t\r\n") {
		return errors.New("execution environment image must be canonical")
	}
	name, digest, ok := strings.Cut(e.Image, "@sha256:")
	if !ok || name == "" || !executionEnvironmentDigest.MatchString(digest) {
		return errors.New("execution environment image must use a sha256 digest")
	}
	if strings.HasPrefix(name, "-") || strings.Contains(name, "://") || strings.Contains(name, "@") || strings.ContainsAny(name, "\\?#") || strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") || strings.Contains(name, "//") {
		return errors.New("invalid OCI image reference")
	}
	switch e.PullPolicy {
	case ExecutionEnvironmentPullAlways, ExecutionEnvironmentPullIfNotPresent, ExecutionEnvironmentPullNever:
		return nil
	default:
		return fmt.Errorf("unsupported execution environment pull policy %q", e.PullPolicy)
	}
}
