package hafence

import (
 "context"
 "database/sql"
 "errors"
)

// ErrAdapterUnavailable is returned when the Semaphore task/outbox transaction
// adapter has not been configured. Never fall back to an unfenced write.
var ErrAdapterUnavailable = errors.New("HA transactional task adapter unavailable")

// TerminalMutation carries the exact durable effects of a runner completion.
// Apply must use the supplied transaction for *every* write. No network I/O,
// notifications or webhook delivery may happen inside Apply.
type TerminalMutation struct {
 Claim Claim
 Apply func(context.Context, *sql.Tx) error
}

// CommitTerminal validates a claim and runs the caller's persistence work
// under the same PostgreSQL transaction and row locks. This primitive does not
// itself update Semaphore's task table: callers must supply a real adapter.
func CommitTerminal(ctx context.Context, db *sql.DB, mutation TerminalMutation) error {
 if mutation.Apply == nil { return ErrAdapterUnavailable }
 return WithClaimLock(ctx, db, mutation.Claim, func(tx *sql.Tx) error {
  return mutation.Apply(ctx, tx)
 })
}
