package tasks

import (
    "errors"
    "testing"
)

func TestTaskPoolDispatchGuardInjection(t *testing.T) {
    p := &TaskPool{}
    if p.dispatchGuard != nil {
        t.Fatal("legacy mode should have no dispatch guard")
    }
    sentinel := errors.New("coordination unavailable")
    p.SetDispatchGuard(func() error { return sentinel })
    if err := p.dispatchGuard(); !errors.Is(err, sentinel) {
        t.Fatalf("guard error = %v; want %v", err, sentinel)
    }
    p.SetDispatchGuard(nil)
    if p.dispatchGuard != nil {
        t.Fatal("expected guard to be cleared")
    }
}
