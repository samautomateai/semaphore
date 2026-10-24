package tasks

import (
    "context"
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

func TestTaskPoolDurableClaimInjection(t *testing.T) {
    p := &TaskPool{}
    if p.claimTask != nil {
        t.Fatal("legacy mode should not have a durable claim callback")
    }
    p.SetClaimTask(func(_ context.Context, id int, token string) (bool, error) {
        return id == 42 && token == "claim-42", nil
    })
    ok, err := p.claimTask(context.Background(), 42)
    if err != nil || !ok {
        t.Fatalf("claim result = (%v, %v)", ok, err)
    }
}

func TestRemoteFinalizeAuthorizerInjection(t *testing.T) {
    p := &TaskPool{}
    p.SetRemoteFinalizeAuthorizer(func(_ context.Context, id int) (bool, error) {
        return id == 42, nil
    })
    allowed, err := p.authorizeRemoteFinalize(context.Background(), 42, "claim-42")
    if err != nil || !allowed {
        t.Fatalf("authorization result = (%v, %v)", allowed, err)
    }
    allowed, err = p.authorizeRemoteFinalize(context.Background(), 43, "claim-42")
    if err != nil || allowed {
        t.Fatalf("unexpected authorization result = (%v, %v)", allowed, err)
    }
}

func TestLegacyRemoteFinalizeRejectedInHAMode(t *testing.T) {
    p := &TaskPool{}
    p.SetClaimTask(func(context.Context, int) (bool, error) { return true, nil })
    // Must reject before accessing state or performing any side effects.
    p.FinalizeRemoteTask(nil, nil)
}
