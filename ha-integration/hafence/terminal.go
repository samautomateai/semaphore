package hafence

import (
 "context"
 "database/sql"
 "errors"
 "fmt"
)

var ErrStaleClaim = errors.New("stale or expired HA claim")

// Claim identifies one specific execution attempt. Tokens must originate from
// the server-side durable claim and be sent to the authenticated runner.
type Claim struct {
 TaskID int64
 Token int64
 OwnerNode string
 OwnerGeneration string
}

// UpdateTerminal atomically validates the active claim and writes a terminal
// status. The task's terminal status must not be updated separately by a
// preceding unfenced handler.
func UpdateTerminal(ctx context.Context, db *sql.DB, c Claim, status string) error {
 if db == nil || c.TaskID <= 0 || c.Token <= 0 || c.OwnerNode == "" || c.OwnerGeneration == "" {
  return errors.New("invalid HA claim")
 }
 switch status {
 case "success", "error", "stopped":
 default:
  return fmt.Errorf("unsupported terminal status %q", status)
 }
 tx, err := db.BeginTx(ctx, nil)
 if err != nil { return err }
 defer tx.Rollback()
 var valid bool
 err = tx.QueryRowContext(ctx, `
 SELECT EXISTS (
  SELECT 1 FROM ha_task_runtime r
  JOIN ha_nodes n ON n.node_id = r.owner_node AND n.generation = r.owner_generation
  WHERE r.task_id = $1 AND r.fencing_token = $2
    AND r.owner_node = $3 AND r.owner_generation = $4
    AND r.lease_expires_at > NOW() AND n.lease_expires_at > NOW()
    AND r.dispatch_phase = 'dispatched'
 )`, c.TaskID, c.Token, c.OwnerNode, c.OwnerGeneration).Scan(&valid)
 if err != nil { return err }
 if !valid { return ErrStaleClaim }
 // This module deliberately does not guess Semaphore's task-status column or
 // enum values. An adapter must implement its update inside THIS transaction.
 return errors.New("Semaphore terminal-status transactional adapter not installed")
}
