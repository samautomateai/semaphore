package hafence

import (
 "context"
 "database/sql"
 "encoding/json"
 "errors"
)

var ErrInvalidTerminal = errors.New("invalid HA terminal event")

// TerminalEvent is written only after the fenced task mutation succeeds.
// A delivery worker must process the outbox after commit and use idempotency.
type TerminalEvent struct {
 Mutation TaskMutation
 Payload json.RawMessage
}

// CommitTaskTerminal atomically updates the Semaphore task and inserts an
// idempotent completion event under the durable claim lock. No external
// effects are emitted here. The event payload must be constructed by trusted
// server code, not passed through unvalidated from the runner.
func CommitTaskTerminal(ctx context.Context, db *sql.DB, e TerminalEvent) error {
 switch e.Mutation.Status {
 case "success", "error", "stopped":
 default:
  return ErrInvalidTerminal
 }
 if !json.Valid(e.Payload) || len(e.Payload) == 0 {
  return ErrInvalidTerminal
 }
 return CommitTerminal(ctx, db, TerminalMutation{
  Claim: e.Mutation.Claim,
  Apply: func(ctx context.Context, tx *sql.Tx) error {
   if err := ApplyTaskMutation(ctx, tx, e.Mutation); err != nil { return err }
   result, err := tx.ExecContext(ctx, `INSERT INTO ha_task_outbox
    (task_id, fencing_token, event_type, payload)
    VALUES ($1, $2, 'task.completed', $3::jsonb)
    ON CONFLICT (task_id, fencing_token, event_type) DO NOTHING`,
    e.Mutation.Claim.TaskID, e.Mutation.Claim.Token, string(e.Payload))
   if err != nil { return err }
   count, err := result.RowsAffected()
   if err != nil { return err }
   if count != 1 { return ErrStaleClaim }
   return nil
  },
 })
}
