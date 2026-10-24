package hafence

import (
 "context"
 "encoding/json"
 "errors"
 "testing"
)

func TestCommitTaskTerminalRejectsNonterminalStatus(t *testing.T) {
 e := TerminalEvent{Mutation: TaskMutation{Status: "running"}, Payload: json.RawMessage(`{"task_id":1}`)}
 if err := CommitTaskTerminal(context.Background(), nil, e); !errors.Is(err, ErrInvalidTerminal) {
  t.Fatalf("expected ErrInvalidTerminal, got %v",err)
 }
}

func TestCommitTaskTerminalRejectsInvalidJSON(t *testing.T) {
 for _, payload := range []json.RawMessage{nil, json.RawMessage("not-json"), json.RawMessage("{")} {
  e := TerminalEvent{Mutation: TaskMutation{Status: "success"}, Payload: payload}
  if err := CommitTaskTerminal(context.Background(), nil, e); !errors.Is(err, ErrInvalidTerminal) {
   t.Fatalf("expected ErrInvalidTerminal for %q, got %v",payload,err)
  }
 }
}
