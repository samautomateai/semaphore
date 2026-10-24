package hafence

import (
 "context"
 "database/sql"
 "errors"
 "fmt"
 "time"
)

// OutboxHandler must be idempotent by record.ID. Delivery is at-least-once.
type OutboxHandler func(context.Context, OutboxRecord) error

type OutboxWorker struct {
 DB *sql.DB
 Owner string
 Handler OutboxHandler
 BatchSize int
 Lease time.Duration
 MaxAttempts int
 RetryDelay time.Duration
}

// RunOnce claims a batch, delivers each event, and acknowledges or schedules
// retry. A handler error is recorded durably; a database error is returned.
// Caller controls polling and shutdown via context.
func (w OutboxWorker) RunOnce(ctx context.Context) (int,error) {
 if w.Handler == nil || w.MaxAttempts <= 0 || w.RetryDelay < 0 {
  return 0,errors.New("invalid HA outbox worker")
 }
 records,err:=ClaimOutbox(ctx,w.DB,w.Owner,w.BatchSize,w.Lease)
 if err!=nil{return 0,err}
 processed:=0
 for _,record:=range records {
  if err:=ctx.Err();err!=nil{return processed,err}
  // Unknown event types must never be passed to a generic handler.
  var deliveryErr error
  if record.EventType!="task.completed" {
   deliveryErr=fmt.Errorf("unsupported HA outbox event type %q",record.EventType)
  } else {
   deliveryErr=w.Handler(ctx,record)
  }
  if deliveryErr!=nil {
   if err:=FailOutbox(ctx,w.DB,record,w.Owner,deliveryErr,w.MaxAttempts,w.RetryDelay);err!=nil {
    return processed,err
   }
  } else {
   if err:=AcknowledgeOutbox(ctx,w.DB,record.ID,w.Owner,record.Attempts);err!=nil {
    return processed,err
   }
  }
  processed++
 }
 return processed,nil
}
