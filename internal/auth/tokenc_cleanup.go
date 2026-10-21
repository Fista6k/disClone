package auth

import (
	"context"
	"log/slog"
	"time"

	"github.com/Fista6k/disClone/internal"
)

func StartTokenCleanup(ctx context.Context, interval time.Duration, storage *internal.Storage) {
	go func() {
		cleanupExpiredTokens(ctx, storage)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cleanupExpiredTokens(ctx, storage)
			}
		}
	}()
}

func cleanupExpiredTokens(ctx context.Context, storage *internal.Storage) {
	cleanupCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	query := `
		DELETE FROM refresh_tokens
		WHERE expired_at <= NOW() OR is_revoked = TRUE;
	`

	result, err := storage.DB.ExecContext(cleanupCtx, query)
	if err != nil {
		slog.Error("failed to clean expired refresh tokens",
			"error", err,
		)
		return
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		slog.Error("failed to get deleted token count",
			"error", err,
		)
		return
	}

	slog.Info("expired refresh tokens cleaned",
		"deleted", deleted,
	)
}
