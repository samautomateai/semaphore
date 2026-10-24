package hafence

import (
 "context"
 "database/sql"
 "errors"
 "testing"
 "time"
)

func TestClaimOutboxRejectsInvalidParameters(t *testing.T) {
 ctx:=context.Background()
 for _, tc:=range []struct{owner string; limit int; lease time.Duration}{
  {"",1,time.Minute},
  {"worker",0,time.Minute},
  {"worker",101,time.Minute},
  {"worker",1,0},
  {"worker",1,2*time.Hour},
 } {
  if _,err:=ClaimOutbox(ctx,(*sql.DB)(nil),tc.owner,tc.limit,tc.lease);err==nil {
   t.Fatalf("accepted invalid parameters: %+v",tc)
  }
 }
}

func TestAcknowledgeOutboxRejectsInvalidParameters(t *testing.T) {
 ctx:=context.Background()
 for _,tc:=range []struct{id int64; owner string; attempt int}{
  {0,"worker",1},
  {1,"",1},
  {1,"worker",0},
 } {
  if err:=AcknowledgeOutbox(ctx,(*sql.DB)(nil),tc.id,tc.owner,tc.attempt);err==nil {
   t.Fatalf("accepted invalid acknowledgment: %+v",tc)
  }
 }
}

func TestOutboxLeaseLostSentinel(t *testing.T) {
 if !errors.Is(ErrOutboxLeaseLost,ErrOutboxLeaseLost) { t.Fatal("missing sentinel") }
}
