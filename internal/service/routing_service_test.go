package service

import (
	"testing"

	"github.com/ghstouch/one-go/internal/model"
)

func TestSelectTarget_Priority(t *testing.T) {
	svc := NewRoutingService()
	combo := &model.Combo{
		Strategy: model.StrategyPriority,
		Targets: []model.ComboTarget{
			{ID: "t1", IsActive: true, Priority: 2},
			{ID: "t2", IsActive: true, Priority: 1},
			{ID: "t3", IsActive: true, Priority: 3},
		},
	}

	target, err := svc.SelectTarget(combo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if target.ID != "t2" {
		t.Fatalf("expected target t2 (lowest priority), got %s", target.ID)
	}
}

func TestSelectTarget_RoundRobin(t *testing.T) {
	svc := NewRoutingService()
	combo := &model.Combo{
		ID:       "combo1",
		Strategy: model.StrategyRoundRobin,
		Targets: []model.ComboTarget{
			{ID: "t1", IsActive: true},
			{ID: "t2", IsActive: true},
			{ID: "t3", IsActive: true},
		},
	}

	results := make(map[string]int)
	for i := 0; i < 9; i++ {
		target, err := svc.SelectTarget(combo)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		results[target.ID]++
	}

	// Each target should be selected 3 times
	for _, id := range []string{"t1", "t2", "t3"} {
		if results[id] != 3 {
			t.Fatalf("expected target %s selected 3 times, got %d", id, results[id])
		}
	}
}

func TestSelectTarget_Weighted(t *testing.T) {
	svc := NewRoutingService()
	combo := &model.Combo{
		Strategy: model.StrategyWeighted,
		Targets: []model.ComboTarget{
			{ID: "t1", IsActive: true, Weight: 100},
			{ID: "t2", IsActive: true, Weight: 0},
		},
	}

	// With weight 100 vs 0, t1 should always be selected
	for i := 0; i < 10; i++ {
		target, err := svc.SelectTarget(combo)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if target.ID != "t1" {
			t.Fatalf("expected t1 with weight 100, got %s", target.ID)
		}
	}
}

func TestSelectTarget_Fallback(t *testing.T) {
	svc := NewRoutingService()
	combo := &model.Combo{
		Strategy: model.StrategyFallback,
		Targets: []model.ComboTarget{
			{ID: "t2", IsActive: true, Priority: 2},
			{ID: "t1", IsActive: true, Priority: 1},
		},
	}

	target, err := svc.SelectTarget(combo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if target.ID != "t1" {
		t.Fatalf("expected t1 (lowest priority), got %s", target.ID)
	}
}

func TestSelectTarget_NilCombo(t *testing.T) {
	svc := NewRoutingService()
	_, err := svc.SelectTarget(nil)
	if err == nil {
		t.Fatal("expected error for nil combo")
	}
}

func TestSelectTarget_NoActiveTargets(t *testing.T) {
	svc := NewRoutingService()
	combo := &model.Combo{
		Strategy: model.StrategyPriority,
		Targets: []model.ComboTarget{
			{ID: "t1", IsActive: false},
		},
	}

	_, err := svc.SelectTarget(combo)
	if err == nil {
		t.Fatal("expected error for no active targets")
	}
}

func TestSelectTarget_SkipsInactive(t *testing.T) {
	svc := NewRoutingService()
	combo := &model.Combo{
		Strategy: model.StrategyPriority,
		Targets: []model.ComboTarget{
			{ID: "t1", IsActive: false, Priority: 1},
			{ID: "t2", IsActive: true, Priority: 2},
		},
	}

	target, err := svc.SelectTarget(combo)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if target.ID != "t2" {
		t.Fatalf("expected t2 (only active), got %s", target.ID)
	}
}

func TestGetFallbackOrder(t *testing.T) {
	combo := &model.Combo{
		Targets: []model.ComboTarget{
			{ID: "t3", IsActive: true, Priority: 3},
			{ID: "t1", IsActive: true, Priority: 1},
			{ID: "t2", IsActive: false, Priority: 2},
			{ID: "t4", IsActive: true, Priority: 2},
		},
	}

	order := GetFallbackOrder(combo)
	if len(order) != 3 {
		t.Fatalf("expected 3 active targets, got %d", len(order))
	}
	if order[0].ID != "t1" || order[1].ID != "t4" || order[2].ID != "t3" {
		t.Fatalf("unexpected order: %s, %s, %s", order[0].ID, order[1].ID, order[2].ID)
	}
}
