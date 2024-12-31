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
		case "dailyCostReport":
			if err := job.DailyCostReport(ctx); err != nil {
				slog.ErrorContext(ctx, "dailyCostReport job was failed", slog.String("error", err.Error()))
				return err
			}

		case "weeklyCostReport":
			if err := job.WeeklyCostReport(ctx); err != nil {
				slog.ErrorContext(ctx, "WeeklyCostReport job failed", slog.String("error", err.Error()))
				return err
			}

		default:
			slog.DebugContext(ctx, "skip to process", slog.String("type:", event.Type))
		}

		return nil
	}
}
