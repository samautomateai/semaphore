package hafence

import (
 "context"
 "database/sql"
 "encoding/json"
 "errors"
 "fmt"
 "time"
)

var ErrOutboxLeaseLost = errors.New("HA outbox delivery lease lost")

type OutboxRecord struct {
 ID int64
 TaskID int64
 FencingToken int64
 EventType string
 Payload json.RawMessage
 Attempts int
}

// ClaimOutbox leases up to limit pending events to a worker. SKIP LOCKED
// supports concurrent workers. All timestamps are evaluated in PostgreSQL.
func ClaimOutbox(ctx context.Context, db *sql.DB, owner string, limit int, lease time.Duration) ([]OutboxRecord, error) {
 if db == nil || owner == "" || limit <= 0 || limit > 100 || lease < time.Second || lease > time.Hour {
  return nil, errors.New("invalid HA outbox lease parameters")
 }
 tx, err := db.BeginTx(ctx, nil)
 if err != nil { return nil, err }
 defer tx.Rollback()
 rows, err := tx.QueryContext(ctx, `
 WITH candidates AS (
  SELECT id FROM ha_task_outbox
  WHERE delivered_at IS NULL
    AND dead_lettered_at IS NULL
    AND available_at <= clock_timestamp()
    AND (delivery_lease_expires_at IS NULL OR delivery_lease_expires_at < clock_timestamp())
  ORDER BY id
  FOR UPDATE SKIP LOCKED LIMIT $1
 )
 UPDATE ha_task_outbox o
 SET delivery_owner=$2,
     delivery_lease_expires_at=clock_timestamp() + ($3 * interval '1 second'),
     attempts=o.attempts+1
 FROM candidates c WHERE o.id=c.id
 RETURNING o.id,o.task_id,o.fencing_token,o.event_type,o.payload,o.attempts
 `, limit, owner, lease.Seconds())
 if err != nil { return nil, err }
 var records []OutboxRecord
 for rows.Next() {
  var rec OutboxRecord
  if err := rows.Scan(&rec.ID,&rec.TaskID,&rec.FencingToken,&rec.EventType,&rec.Payload,&rec.Attempts); err != nil {
   rows.Close()
   return nil, err
  }
  records=append(records,rec)
 }
 if err := rows.Err(); err != nil { rows.Close(); return nil, err }
 if err := rows.Close(); err != nil { return nil, err }
 if err := tx.Commit(); err != nil { return nil, err }
 return records,nil
}

// AcknowledgeOutbox marks a delivered event only while the same worker owns
// an unexpired lease and matching attempt number. A stale worker cannot
// acknowledge a reassigned event even if the same owner name is reused.
// The downstream consumer must deduplicate by outbox ID, since a crash after
// delivery but before this acknowledgment will cause redelivery.
func AcknowledgeOutbox(ctx context.Context, db *sql.DB, id int64, owner string, attempt int) error {
 if db == nil || id <= 0 || owner == "" || attempt <= 0 { return errors.New("invalid HA outbox acknowledgment") }
 result, err := db.ExecContext(ctx, `
 UPDATE ha_task_outbox SET delivered_at=clock_timestamp(),
 delivery_owner=NULL, delivery_lease_expires_at=NULL
 WHERE id=$1 AND delivery_owner=$2 AND attempts=$3 AND delivered_at IS NULL
 AND delivery_lease_expires_at > clock_timestamp()
 `,id,owner,attempt)
 if err != nil { return err }
 n,err:=result.RowsAffected()
 if err != nil { return err }
 if n!=1 { return fmt.Errorf("%w: event %d",ErrOutboxLeaseLost,id) }
 return nil
}
