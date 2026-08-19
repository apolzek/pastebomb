package bootstrap

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/apolzek/pastebomb/internal/application"
	"github.com/apolzek/pastebomb/internal/infrastructure/config"
)

// PurgeWorker deletes expired pastes on a schedule. Reads already hide expired
// pastes, so this job exists to reclaim storage, not to enforce the rule.
type PurgeWorker struct {
	pastes *application.PasteService
	cfg    config.Purge
	done   sync.WaitGroup
}

func NewPurgeWorker(pastes *application.PasteService, cfg config.Purge) *PurgeWorker {
	return &PurgeWorker{pastes: pastes, cfg: cfg}
}

// Start runs the sweep loop until ctx is cancelled.
func (w *PurgeWorker) Start(ctx context.Context) {
	w.done.Add(1)
	go func() {
		defer w.done.Done()

		ticker := time.NewTicker(w.cfg.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Info("purge worker stopped")
				return
			case <-ticker.C:
				w.sweep(ctx)
			}
		}
	}()
}

// Wait blocks until the loop has exited.
func (w *PurgeWorker) Wait() { w.done.Wait() }

// sweep runs one pass under its own timeout, so a slow delete cannot stall the
// ticker indefinitely. A failure is logged and retried on the next tick.
func (w *PurgeWorker) sweep(ctx context.Context) {
	sweepCtx, cancel := context.WithTimeout(ctx, w.cfg.Timeout)
	defer cancel()

	deleted, err := w.pastes.PurgeExpired(sweepCtx)
	if err != nil {
		slog.Error("purge of expired pastes failed", "err", err)
		return
	}
	if deleted > 0 {
		slog.Info("expired pastes purged", "count", deleted)
	}
}
