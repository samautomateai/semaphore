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
    p.SetClaimTask(func(_ context.Context, id int) (*DurableTaskClaim, error) {
        if id != 42 { return nil, nil }
        return &DurableTaskClaim{TaskID: id, Token: "claim-42", OwnerNode: "node-a", OwnerGeneration: "generation-a"}, nil
    })
    claim, err := p.claimTask(context.Background(), 42)
    if err != nil || claim == nil || claim.Token != "claim-42" {
        t.Fatalf("claim result = (%+v, %v)", claim, err)
    }
}

func TestRemoteFinalizeAuthorizerInjection(t *testing.T) {
    p := &TaskPool{}
    p.SetRemoteFinalizeAuthorizer(func(_ context.Context, id int, token string) (bool, error) {
        return id == 42 && token == "claim-42", nil
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
    p.SetClaimTask(func(context.Context, int) (*DurableTaskClaim, error) { return &DurableTaskClaim{TaskID: 1, Token: "t", OwnerNode: "n", OwnerGeneration: "g"}, nil })
    // Must reject before accessing state or performing any side effects.
    p.FinalizeRemoteTask(nil, nil)
}

func TestTaskRunnerDurableClaimReturnsCopy(t *testing.T) {
    runner := &TaskRunner{durableClaim: &DurableTaskClaim{
        TaskID: 42, Token: "token-42", OwnerNode: "node-a", OwnerGeneration: "generation-a",
    }}
    copy := runner.DurableClaim()
    if copy == nil || copy.Token != "token-42" {
        t.Fatalf("unexpected claim: %+v", copy)
    }
    copy.Token = "tampered"
    if runner.DurableClaim().Token != "token-42" {
        t.Fatal("caller mutated original durable claim")
    }
}

func TestTaskRunnerDurableClaimLegacyMode(t *testing.T) {
    runner := &TaskRunner{}
    if runner.DurableClaim() != nil {
        t.Fatal("legacy runner should have no durable claim")
    }
}
