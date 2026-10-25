package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunnerHeartbeatEligibility(t *testing.T) {
	now := time.Date(2026, 10, 25, 13, 0, 0, 0, time.UTC)
	limit := 30 * time.Second
	tests := []struct {
		name string
		touched *time.Time
		window time.Duration
		webhook string
		want bool
	}{
		{name:"never polled", window:limit},
		{name:"negative window", touched:timePtr(now), window:-time.Second},
		{name:"zero window", touched:timePtr(now)},
		{name:"future heartbeat", touched:timePtr(now.Add(time.Minute)), window:limit},
		{name:"within timeout", touched:timePtr(now.Add(-time.Second)), window:limit, want:true},
		{name:"exact timeout", touched:timePtr(now.Add(-limit)), window:limit, want:true},
		{name:"expired", touched:timePtr(now.Add(-limit-time.Nanosecond)), window:limit},
		{name:"webhook driven", webhook:"https://runner.example.test", want:true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Runner{Touched:tt.touched, Webhook:tt.webhook}
			require.Equal(t, tt.want, r.IsOnline(now, tt.window))
			r.FillStatus(now, tt.window)
			if tt.want { require.Equal(t, RunnerStatusOnline, r.Status) } else { require.Equal(t, RunnerStatusOffline, r.Status) }
		})
	}
}

func timePtr(t time.Time) *time.Time { return &t }

func TestRunnerDispatchEligibility(t *testing.T) {
	now := time.Date(2026, 10, 25, 13, 0, 0, 0, time.UTC)
	heartbeat := now.Add(-time.Second)
	base := Runner{Active:true, Token:"registered-token", Touched:&heartbeat, MaxParallelTasks:2}
	require.True(t, base.CanAcceptTask(now, 30*time.Second, 0))
	require.True(t, base.CanAcceptTask(now, 30*time.Second, 1))
	require.False(t, base.CanAcceptTask(now, 30*time.Second, 2))
	require.False(t, base.CanAcceptTask(now, 30*time.Second, -1))
	inactive := base
	inactive.Active = false
	require.False(t, inactive.CanAcceptTask(now, 30*time.Second, 0))
	unregistered := base
	unregistered.Token = ""
	require.False(t, unregistered.CanAcceptTask(now, 30*time.Second, 0))
	stale := base
	old := now.Add(-time.Minute)
	stale.Touched = &old
	require.False(t, stale.CanAcceptTask(now, 30*time.Second, 0))
	unlimited := base
	unlimited.MaxParallelTasks = 0
	require.True(t, unlimited.CanAcceptTask(now, 30*time.Second, 1000))
	invalid := base
	invalid.MaxParallelTasks = -1
	require.False(t, invalid.CanAcceptTask(now, 30*time.Second, 0))
}
