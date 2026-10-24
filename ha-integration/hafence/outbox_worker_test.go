package hafence

import (
 "context"
 "errors"
 "testing"
 "time"
)

func TestOutboxWorkerRejectsMissingHandler(t *testing.T) {
 worker := OutboxWorker{Owner:"worker",BatchSize:1,Lease:time.Minute,MaxAttempts:3,RetryDelay:time.Second}
 _,err:=worker.RunOnce(context.Background())
 if err==nil { t.Fatal("missing handler must fail closed") }
}

func TestOutboxWorkerRejectsInvalidPolicy(t *testing.T) {
 handler:=func(context.Context,OutboxRecord)error{return nil}
 for _,tc:=range []OutboxWorker{
  {Handler:handler,MaxAttempts:0},
  {Handler:handler,MaxAttempts:3,RetryDelay:-time.Second},
 } {
  if _,err:=tc.RunOnce(context.Background());err==nil {t.Fatal("invalid policy accepted")}
 }
}

func TestOutboxWorkerCancellation(t *testing.T) {
 ctx,cancel:=context.WithCancel(context.Background())
 cancel()
 handler:=func(context.Context,OutboxRecord)error{return errors.New("must not be called")}
 worker:=OutboxWorker{Handler:handler,Owner:"worker",BatchSize:1,Lease:time.Minute,MaxAttempts:3,RetryDelay:time.Second}
 // ClaimOutbox must refuse a nil DB even if cancellation occurred.
 if _,err:=worker.RunOnce(ctx);err==nil {t.Fatal("missing DB accepted")}
}
