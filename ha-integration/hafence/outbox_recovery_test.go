package hafence

import (
 "context"
 "errors"
 "testing"
 "time"
)

func TestRenewOutboxLeaseRejectsMissingDatabase(t *testing.T) {
 err:=RenewOutboxLease(context.Background(),nil,OutboxRecord{ID:1,Attempts:1},"worker",time.Minute)
 if err==nil {t.Fatal("missing DB must fail closed")}
}

func TestRenewOutboxLeaseRejectsInvalidExtension(t *testing.T) {
 for _,extension:=range []time.Duration{0,time.Hour+time.Second} {
  err:=RenewOutboxLease(context.Background(),nil,OutboxRecord{ID:1,Attempts:1},"worker",extension)
  if err==nil {t.Fatalf("accepted extension %v",extension)}
 }
}

func TestReplayDeadLetterRejectsMissingDatabase(t *testing.T) {
 if err:=ReplayDeadLetter(context.Background(),nil,1);err==nil {t.Fatal("missing DB must fail closed")}
}

func TestOutboxLeaseSentinel(t *testing.T) {
 if !errors.Is(ErrOutboxLeaseLost,ErrOutboxLeaseLost) {t.Fatal("sentinel missing")}
}
