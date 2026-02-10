package state

import (
	"testing"

	"github.com/johnnyfreeman/viewscreen/result"
)

func TestNewState(t *testing.T) {
	s := NewState()
	if s == nil {
		t.Fatal("NewState should return non-nil")
	}
	if s.InputTokens != 0 {
		t.Errorf("InputTokens should be 0, got %d", s.InputTokens)
	}
	if s.OutputTokens != 0 {
		t.Errorf("OutputTokens should be 0, got %d", s.OutputTokens)
	}
}

func TestAccumulateUsage(t *testing.T) {
	s := NewState()

	s.AccumulateUsage(100, 50, 200, 300)
	if s.InputTokens != 100 {
		t.Errorf("InputTokens = %d, want 100", s.InputTokens)
	}
	if s.OutputTokens != 50 {
		t.Errorf("OutputTokens = %d, want 50", s.OutputTokens)
	}
	if s.CacheCreated != 200 {
		t.Errorf("CacheCreated = %d, want 200", s.CacheCreated)
	}
	if s.CacheRead != 300 {
		t.Errorf("CacheRead = %d, want 300", s.CacheRead)
	}

	// Accumulate more
	s.AccumulateUsage(50, 25, 0, 100)
	if s.InputTokens != 150 {
		t.Errorf("InputTokens = %d, want 150", s.InputTokens)
	}
	if s.OutputTokens != 75 {
		t.Errorf("OutputTokens = %d, want 75", s.OutputTokens)
	}
	if s.CacheCreated != 200 {
		t.Errorf("CacheCreated = %d, want 200", s.CacheCreated)
	}
	if s.CacheRead != 400 {
		t.Errorf("CacheRead = %d, want 400", s.CacheRead)
	}
}

func TestAccumulateUsage_ThenResultOverrides(t *testing.T) {
	s := NewState()

	// Accumulate from per-turn usage
	s.AccumulateUsage(100, 50, 200, 300)

	// Result event should override with authoritative totals
	s.UpdateFromResultEvent(result.Event{
		NumTurns:     5,
		TotalCostUSD: 0.05,
		Usage: result.Usage{
			InputTokens:              500,
			OutputTokens:             200,
			CacheCreationInputTokens: 1000,
			CacheReadInputTokens:     2000,
		},
	})

	if s.InputTokens != 500 {
		t.Errorf("InputTokens = %d, want 500 (should be overridden by result)", s.InputTokens)
	}
	if s.OutputTokens != 200 {
		t.Errorf("OutputTokens = %d, want 200 (should be overridden by result)", s.OutputTokens)
	}
}
