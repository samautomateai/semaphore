package hafence

import (
 "context"
 "errors"
 "testing"
)

func TestWithClaimLockRejectsInvalidInputs(t *testing.T) {
 err := WithClaimLock(context.Background(),nil,Claim{},nil)
 if err == nil { t.Fatal("expected invalid claim error") }
}

func TestUpdateTerminalFailsClosed(t *testing.T) {
 err := UpdateTerminal(context.Background(),nil,Claim{},"success")
 if err == nil || errors.Is(err,ErrStaleClaim) { t.Fatalf("unexpected error: %v",err) }
}
