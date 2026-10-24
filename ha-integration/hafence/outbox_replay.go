package hafence

import (
 "context"
 "database/sql"
 "errors"
 "fmt"
)

// ReplayDeadLetter makes an operator-selected dead-letter event eligible
// for redelivery. This does not reset the attempt count, preserving audit
// history; callers should configure MaxAttempts above the existing count.
// Do not expose this without operator authorization and an audit record.
func ReplayDeadLetter(ctx context.Context, db *sql.DB, id int64) error {
 if db == nil || id <= 0 { return errors.New("invalid HA dead-letter replay") }
 result,err:=db.ExecContext(ctx, `
 UPDATE ha_task_outbox
 SET dead_lettered_at=NULL, last_error=NULL,
  available_at=clock_timestamp(), delivery_owner=NULL,
  delivery_lease_expires_at=NULL
 WHERE id=$1 AND delivered_at IS NULL AND dead_lettered_at IS NOT NULL
 `,id)
 if err!=nil{return err}
 n,err:=result.RowsAffected()
 if err!=nil{return err}
 if n!=1{return fmt.Errorf("HA dead-letter event %d not replayable",id)}
 return nil
}
