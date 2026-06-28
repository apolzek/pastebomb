package bootstrap

import (
	"log/slog"
	"time"

	"gin-goinc-api/repository"
)

// startExpiredPostPurger walks the posts table on an interval and deletes
// rows whose expiration_date has passed. Runs forever in its own goroutine.
// The DB query is bounded by the indexed expiration_date column.
func startExpiredPostPurger(interval time.Duration) {
	go func() {
		repo := repository.NewPostRepository()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			n, err := repo.PurgeExpiredPosts()
			if err != nil {
				slog.Error("purge expired posts failed", "err", err)
				continue
			}
			if n > 0 {
				slog.Info("purged expired posts", "count", n)
			}
		}
	}()
}
