package hafence

import (
 "context"
 "database/sql"
 "errors"
 "testing"
)

func TestApplyTaskMutationRejectsInvalidInput(t *testing.T) {
 cases := []TaskMutation{
  {},
  {Claim: Claim{TaskID: 1}, RunnerID: 1, Status: "success"},
  {Claim: Claim{TaskID: 1}, RunnerID: 0, Status: "running"},
 }
 for i, c := range cases {
  // A nil transaction must always fail closed, even for otherwise valid data.
  if err := ApplyTaskMutation(context.Background(), (*sql.Tx)(nil), c); !errors.Is(err, ErrInvalidMutation) {
   t.Fatalf("case %d: expected invalid mutation, got %v", i, err)
  }
 }
}

func TestCommitTaskMutationRejectsMissingDatabase(t *testing.T) {
 m := TaskMutation{Claim: Claim{TaskID: 1, Token: 1, OwnerNode: "node", OwnerGeneration: "generation"}, RunnerID: 1, Status: "success"}
 if err := CommitTaskMutation(context.Background(), nil, m); err == nil {
  t.Fatal("missing database must fail closed")
 }
}
