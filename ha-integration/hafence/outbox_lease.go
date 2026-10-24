package hafence

import (
 "context"
 "database/sql"
 "errors"
 "fmt"
 "time"
)

// RenewOutboxLease extends a worker's lease only for the same claim attempt.
// Renew before expiry; callers must stop delivery if this returns lease lost.
func RenewOutboxLease(ctx context.Context, db *sql.DB, record OutboxRecord, owner string, extension time.Duration) error {
 if db == nil || record.ID <= 0 || record.Attempts <= 0 || owner == "" ||
  extension < time.Second || extension > time.Hour {
  return errors.New("invalid HA outbox lease renewal")
 }
 result,err:=db.ExecContext(ctx, `
 UPDATE ha_task_outbox
 SET delivery_lease_expires_at=clock_timestamp()+($4 * interval '1 second')
 WHERE id=$1 AND delivery_owner=$2 AND attempts=$3
 AND delivered_at IS NULL AND dead_lettered_at IS NULL
 AND delivery_lease_expires_at > clock_timestamp()
 `,record.ID,owner,record.Attempts,extension.Seconds())
 if err!=nil{return err}
 n,err:=result.RowsAffected()
 if err!=nil{return err}
 if n!=1{return fmt.Errorf("%w: event %d",ErrOutboxLeaseLost,record.ID)}
 return nil
}
