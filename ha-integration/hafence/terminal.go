package hafence

import (
 "context"
 "database/sql"
 "errors"
 "fmt"
)

var ErrStaleClaim = errors.New("stale or expired HA claim")

// Claim identifies a single durable execution attempt.
type Claim struct {
 TaskID int64
 Token int64
 OwnerNode string
 OwnerGeneration string
}

// WithClaimLock holds the PostgreSQL claim row lock while applying a
// Semaphore-specific status/output update in the SAME transaction.
// The callback must not commit, and must use tx (not db) for all writes.
// Side effects such as webhooks must occur after commit through an outbox.
func WithClaimLock(ctx context.Context, db *sql.DB, c Claim, apply func(*sql.Tx) error) error {
 if db == nil || c.TaskID <= 0 || c.Token <= 0 || c.OwnerNode == "" || c.OwnerGeneration == "" || apply == nil {
  return errors.New("invalid HA claim or transaction callback")
 }
 tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
 if err != nil { return err }
 defer tx.Rollback()
 // FOR UPDATE serializes ownership changes against fenced writes. The node
 // row is also locked to serialize generation changes and heartbeat expiry.
 var token int64
 var phase string
 var valid bool
 err = tx.QueryRowContext(ctx, `
 SELECT q.fencing_token, q.dispatch_phase,
   (q.state='claimed' AND q.lease_expires_at > clock_timestamp()
    AND n.lease_expires_at > clock_timestamp())
 FROM ha_task_queue q
 JOIN ha_nodes n ON n.node_id=q.owner_node AND n.generation=q.owner_generation
 WHERE q.task_id=$1 AND q.fencing_token=$2 AND q.owner_node=$3
   AND q.owner_generation=$4::uuid
 FOR UPDATE OF q,n
 `, c.TaskID,c.Token,c.OwnerNode,c.OwnerGeneration).Scan(&token,&phase,&valid)
 if errors.Is(err, sql.ErrNoRows) { return ErrStaleClaim }
 if err != nil { return err }
 if !valid || token != c.Token || phase != "dispatched" { return ErrStaleClaim }
 if err := apply(tx); err != nil { return fmt.Errorf("fenced write: %w",err) }
 // A long-running callback must not outlive its lease.
 var stillValid bool
 err = tx.QueryRowContext(ctx, `SELECT lease_expires_at > clock_timestamp()
 FROM ha_task_queue WHERE task_id=$1 AND fencing_token=$2`, c.TaskID,c.Token).Scan(&stillValid)
 if err != nil { return err }
 if !stillValid { return ErrStaleClaim }
 return tx.Commit()
}

// UpdateTerminal is deliberately unavailable until the Semaphore DB adapter
// can atomically write its real task status and durable outbox inside the
// transaction. Never silently report a successful completion.
func UpdateTerminal(ctx context.Context, db *sql.DB, c Claim, status string) error {
 return fmt.Errorf("terminal status %q: Semaphore transactional adapter not installed",status)
}
