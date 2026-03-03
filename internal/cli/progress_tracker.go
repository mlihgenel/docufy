package cli

import (
	"sync"
	"time"

	"github.com/mlihgenel/docufy/v2/internal/converter"
)

type progressTracker struct {
	mu        sync.RWMutex
	startedAt time.Time
	percent   float64
	label     string
	current   time.Duration
	total     time.Duration
	eta       time.Duration
	completed int
	totalJobs int
}

func newProgressTracker() *progressTracker {
	return &progressTracker{startedAt: time.Now()}
}

func (p *progressTracker) Update(info converter.ProgressInfo) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if info.Percent > p.percent {
		p.percent = info.Percent
	} else if info.Percent == 100 {
		p.percent = 100
	}
	if info.CurrentLabel != "" {
		p.label = info.CurrentLabel
	}
	if info.Current > 0 {
		p.current = info.Current
	}
	if info.Total > 0 {
		p.total = info.Total
	}
	if info.ETA >= 0 {
		p.eta = info.ETA
	}
	if info.Completed > 0 {
		p.completed = info.Completed
	}
	if info.TotalItems > 0 {
		p.totalJobs = info.TotalItems
	}
}

func (p *progressTracker) Snapshot() converter.ProgressInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return converter.ProgressInfo{
		Percent:      p.percent,
		Current:      p.current,
		Total:        p.total,
		ETA:          p.eta,
		Completed:    p.completed,
		TotalItems:   p.totalJobs,
		CurrentLabel: p.label,
	}
}
