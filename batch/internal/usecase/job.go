package usecase

import (
	"context"

	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
)

type Jobber interface {
	NotifyDaily(ctx context.Context) error
}

var _ Jobber = (*Job)(nil)

type Job struct {
}

func NewJob(cfg configuration.Config) (*Job, error) {

	return &Job{}, nil
}
