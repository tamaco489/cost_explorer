package handler

import (
	"context"
	"log/slog"

	"github.com/tamaco489/cost_explorer/batch/internal/usecase"
)

type Job func(ctx context.Context, event JobEvent) error

func JobHandler(job usecase.Job) Job {
	return func(ctx context.Context, event JobEvent) error {

		switch event.Type {
		case "notifyDaily":
			if err := job.NotifyDaily(ctx); err != nil {
				slog.ErrorContext(ctx, "notifyDaily job was failed", slog.String("error", err.Error()))
				return err
			}

		default:
			slog.DebugContext(ctx, "skip to process", slog.String("type:", event.Type))
		}

		return nil
	}
}
