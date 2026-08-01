// Package run owns background scan scheduling: periodic rescans on a timer,
// plus an on-demand trigger (e.g. a "Rescan now" button in the dashboard).
package run

import (
	"log"
	"sync/atomic"
	"time"

	"nastree/internal/scanner"
	"nastree/internal/store"
)

type Runner struct {
	rootPath string
	interval time.Duration
	retain   int
	store    *store.Store

	running atomic.Bool
	trigger chan struct{}
}

func New(rootPath string, interval time.Duration, retain int, st *store.Store) *Runner {
	return &Runner{
		rootPath: rootPath,
		interval: interval,
		retain:   retain,
		store:    st,
		trigger:  make(chan struct{}, 1),
	}
}

// Trigger requests an immediate scan. Returns false if one is already running.
func (r *Runner) Trigger() bool {
	if r.running.Load() {
		return false
	}
	select {
	case r.trigger <- struct{}{}:
		return true
	default:
		return false
	}
}

func (r *Runner) Running() bool {
	return r.running.Load()
}

// Start runs an initial scan immediately, then reschedules on interval or trigger.
// Blocks until the process exits; run it in a goroutine.
func (r *Runner) Start() {
	r.scanOnce()
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			r.scanOnce()
		case <-r.trigger:
			r.scanOnce()
		}
	}
}

func (r *Runner) scanOnce() {
	if !r.running.CompareAndSwap(false, true) {
		return
	}
	defer r.running.Store(false)

	log.Printf("scan starting: %s", r.rootPath)
	started := time.Now()

	res, err := scanner.Walk(r.rootPath)
	if err != nil {
		log.Printf("scan failed: %v", err)
		return
	}
	finished := time.Now()

	total, free, err := scanner.DiskUsage(r.rootPath)
	if err != nil {
		log.Printf("disk usage lookup failed: %v", err)
	}

	scanID, err := r.store.SaveScan(res, total, free, started, finished, r.retain)
	if err != nil {
		log.Printf("scan save failed: %v", err)
		return
	}
	log.Printf("scan %d complete in %s: %d files, %d dirs, %.2f GB",
		scanID, finished.Sub(started), res.TotalFiles, res.TotalDirs, float64(res.TotalSize)/(1<<30))
}
