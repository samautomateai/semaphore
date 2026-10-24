package hafence

import (
 "errors"
 "testing"
)

func TestAssignmentClaim(t *testing.T) {
 a := Assignment{TaskID: 12, FencingToken: "42", OwnerNode: "node-a", OwnerGeneration: "e56e6aa1-4d36-4e64-9e43-70726e98a700"}
 claim, err := a.Claim()
 if err != nil { t.Fatal(err) }
 if claim.TaskID != 12 || claim.Token != 42 || claim.OwnerNode != a.OwnerNode || claim.OwnerGeneration != a.OwnerGeneration {
  t.Fatalf("unexpected claim: %+v",claim)
 }
}

func TestAssignmentClaimRejectsInvalidTokens(t *testing.T) {
 for _, token := range []string{"", "0", "-1", "not-a-number", "9223372036854775808"} {
  t.Run(token, func(t *testing.T) {
   a := Assignment{TaskID: 12, FencingToken: token, OwnerNode: "node-a", OwnerGeneration: "generation"}
   _, err := a.Claim()
   if !errors.Is(err, ErrInvalidReport) { t.Fatalf("token %q: got %v",token,err) }
  })
 }
}

func TestCompletionValidation(t *testing.T) {
 c := Completion{Assignment: Assignment{TaskID: 1,FencingToken:"3",OwnerNode:"node-a",OwnerGeneration:"generation"},RunnerID:2,Status:"success"}
 if err := c.Validate(); err != nil { t.Fatal(err) }
 c.Status = "running"
 if !errors.Is(c.Validate(), ErrInvalidReport) { t.Fatal("unsupported completion status accepted") }
}
