package hafence

import (
 "errors"
 "fmt"
 "strconv"
)

var ErrInvalidReport = errors.New("invalid fenced runner report")

// Assignment is the server-issued durable identity for one runner execution.
// The token must be transported over an authenticated runner channel.
type Assignment struct {
 TaskID int64 `json:"task_id"`
 FencingToken string `json:"fencing_token"`
 OwnerNode string `json:"owner_node"`
 OwnerGeneration string `json:"owner_generation"`
}

// Completion is the runner's echo of its original assignment identity.
// A server must never populate these fields from the current database owner.
type Completion struct {
 Assignment
 RunnerID int64 `json:"runner_id"`
 Status string `json:"status"`
}

func (a Assignment) Validate() error {
 if a.TaskID <= 0 || a.FencingToken == "" || a.OwnerNode == "" || a.OwnerGeneration == "" {
  return ErrInvalidReport
 }
 return nil
}

func (c Completion) Validate() error {
 if err := c.Assignment.Validate(); err != nil { return err }
 if c.RunnerID <= 0 { return ErrInvalidReport }
 switch c.Status {
 case "success", "error", "stopped":
  return nil
 default:
  return fmt.Errorf("%w: unsupported status %q",ErrInvalidReport,c.Status)
 }
}

// Claim converts a runner's opaque wire identity into the PostgreSQL fencing
// representation. Tokens are positive decimal integers; never silently coerce
// malformed or overflowed values to zero.
func (a Assignment) Claim() (Claim, error) {
 if err := a.Validate(); err != nil { return Claim{}, err }
 token, err := strconv.ParseInt(a.FencingToken, 10, 64)
 if err != nil || token <= 0 {
  return Claim{}, fmt.Errorf("%w: fencing token must be a positive decimal integer", ErrInvalidReport)
 }
 return Claim{
  TaskID: a.TaskID,
  Token: token,
  OwnerNode: a.OwnerNode,
  OwnerGeneration: a.OwnerGeneration,
 }, nil
}
