package hafence

import (
 "context"
 "database/sql"
 "errors"
 "fmt"
)

var ErrInvalidMutation = errors.New("invalid HA task mutation")

// TaskMutation contains the database effects allowed under a valid fencing
// claim. The runner ID must match the server-authorized runner assignment.
type TaskMutation struct {
 Claim Claim
 RunnerID int64
 Status string
 Outputs *string
}

// ApplyTaskMutation updates the real Semaphore task row using the SAME
// transaction that WithClaimLock uses to validate the claim. This helper
// deliberately performs no webhooks, Redis updates or workflow advancement.
// Those effects require a durable outbox in the same transaction.
func ApplyTaskMutation(ctx context.Context, tx *sql.Tx, m TaskMutation) error {
 if tx == nil || m.Claim.TaskID <= 0 || m.RunnerID <= 0 { return ErrInvalidMutation }
 switch m.Status {
 case "running", "success", "error", "stopped":
 default:
  return fmt.Errorf("%w: unsupported status %q", ErrInvalidMutation, m.Status)
 }
 // Do not overwrite artifacts when the runner has not supplied outputs.
 // The runner binding and existing nonterminal status are checked by SQL.
 var result sql.Result
 var err error
 if m.Outputs == nil {
  result, err = tx.ExecContext(ctx, `UPDATE task
 SET status=$1
 WHERE id=$2 AND runner_id=$3 AND status NOT IN ('success','error','stopped')`,
   m.Status, m.Claim.TaskID, m.RunnerID)
 } else {
  result, err = tx.ExecContext(ctx, `UPDATE task
 SET status=$1, artifacts=$2
 WHERE id=$3 AND runner_id=$4 AND status NOT IN ('success','error','stopped')`,
   m.Status, *m.Outputs, m.Claim.TaskID, m.RunnerID)
 }
 if err != nil { return err }
 affected, err := result.RowsAffected()
 if err != nil { return err }
 if affected != 1 { return ErrStaleClaim }
 return nil
}

// CommitTaskMutation requires a real claim lock before applying the mutation.
// It is not a full terminal completion adapter: event/outbox writes are not
// yet implemented and this function must not be wired to production handlers.
func CommitTaskMutation(ctx context.Context, db *sql.DB, m TaskMutation) error {
 return CommitTerminal(ctx, db, TerminalMutation{
  Claim: m.Claim,
  Apply: func(ctx context.Context, tx *sql.Tx) error {
   return ApplyTaskMutation(ctx, tx, m)
  },
 })
}
