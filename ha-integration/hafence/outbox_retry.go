package hafence

import (
 "context"
 "database/sql"
 "errors"
 "fmt"
 "time"
)

const maxOutboxErrorLength = 2048

// FailOutbox schedules a retry or dead-letters an event after maxAttempts.
// A stale worker cannot change the delivery state: owner, attempt and lease
// must all match. maxAttempts counts actual claims, not separate failures.
func FailOutbox(ctx context.Context, db *sql.DB, record OutboxRecord, owner string, cause error, maxAttempts int, retryDelay time.Duration) error {
 if db == nil || record.ID <= 0 || record.Attempts <= 0 || owner == "" || cause == nil ||
  maxAttempts <= 0 || retryDelay < 0 || retryDelay > 24*time.Hour {
  return errors.New("invalid HA outbox failure parameters")
 }
 msg := cause.Error()
 if len(msg) > maxOutboxErrorLength { msg = msg[:maxOutboxErrorLength] }
 deadLetter := record.Attempts >= maxAttempts
 result, err := db.ExecContext(ctx, `
 UPDATE ha_task_outbox SET
  delivery_owner=NULL, delivery_lease_expires_at=NULL,
  last_error=$4,
  available_at=CASE WHEN $5 THEN available_at ELSE clock_timestamp()+($6 * interval '1 second') END,
  dead_lettered_at=CASE WHEN $5 THEN clock_timestamp() ELSE NULL END
 WHERE id=$1 AND delivery_owner=$2 AND attempts=$3 AND delivered_at IS NULL
  AND dead_lettered_at IS NULL AND delivery_lease_expires_at > clock_timestamp()
 `,record.ID,owner,record.Attempts,msg,deadLetter,retryDelay.Seconds())
 if err != nil { return err }
 n,err:=result.RowsAffected()
 if err != nil { return err }
 if n!=1 { return fmt.Errorf("%w: event %d",ErrOutboxLeaseLost,record.ID) }
 return nil
}
