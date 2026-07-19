package service

import (
	"errors"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ghstouch/one-go/internal/model"
)

// RoutingService selects a combo target based on strategy
type RoutingService interface {
	SelectTarget(combo *model.Combo) (*model.ComboTarget, error)
}

type routingService struct {
	rrCounters map[string]*uint64
	mu         sync.RWMutex
}

func NewRoutingService() RoutingService {
	return &routingService{
		rrCounters: make(map[string]*uint64),
	}
}

// SelectTarget picks a target from the combo based on its strategy
func (s *routingService) SelectTarget(combo *model.Combo) (*model.ComboTarget, error) {
	if combo == nil {
		return nil, errors.New("combo is nil")
	}

	// Filter active targets only
	var active []model.ComboTarget
	for _, t := range combo.Targets {
		if t.IsActive {
			active = append(active, t)
		}
	}
	if len(active) == 0 {
		return nil, errors.New("no active targets in combo")
	}

	switch combo.Strategy {
	case model.StrategyPriority:
		return s.selectByPriority(active), nil
	case model.StrategyRoundRobin:
		return s.selectByRoundRobin(combo.ID, active), nil
	case model.StrategyWeighted:
		return s.selectByWeighted(active), nil
	case model.StrategyFallback:
		return s.selectByFallback(active), nil
	default:
		return s.selectByPriority(active), nil
	}
}

// selectByPriority picks the target with the lowest priority number
func (s *routingService) selectByPriority(targets []model.ComboTarget) *model.ComboTarget {
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].Priority < targets[j].Priority
	})
	return &targets[0]
}

// selectByRoundRobin rotates through targets
func (s *routingService) selectByRoundRobin(comboID string, targets []model.ComboTarget) *model.ComboTarget {
	s.mu.Lock()
	counter, exists := s.rrCounters[comboID]
	if !exists {
		var c uint64
		counter = &c
		s.rrCounters[comboID] = counter
	}
	s.mu.Unlock()

	idx := atomic.AddUint64(counter, 1)
	return &targets[idx%uint64(len(targets))]
}

// selectByWeighted picks a target based on weight using weighted random
func (s *routingService) selectByWeighted(targets []model.ComboTarget) *model.ComboTarget {
	totalWeight := 0
	for _, t := range targets {
		w := t.Weight
		if w <= 0 {
			w = 1
		}
		totalWeight += w
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	roll := r.Intn(totalWeight)

	cumulative := 0
	for i, t := range targets {
		w := t.Weight
		if w <= 0 {
			w = 1
		}
		cumulative += w
		if roll < cumulative {
			return &targets[i]
		}
	}
	return &targets[0]
}

// selectByFallback picks the first active target (sorted by priority)
// The actual fallback to next target happens at request time when the first fails
func (s *routingService) selectByFallback(targets []model.ComboTarget) *model.ComboTarget {
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].Priority < targets[j].Priority
	})
	return &targets[0]
}

// GetFallbackOrder returns all targets sorted by priority for fallback retry
func GetFallbackOrder(combo *model.Combo) []model.ComboTarget {
	var active []model.ComboTarget
	for _, t := range combo.Targets {
		if t.IsActive {
			active = append(active, t)
		}
	}
	sort.Slice(active, func(i, j int) bool {
		return active[i].Priority < active[j].Priority
	})
	return active
}
