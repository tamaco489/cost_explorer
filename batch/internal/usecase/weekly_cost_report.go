package usecase

import (
	"context"
	"log"

	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
)

func (j *Job) WeeklyCostReport(ctx context.Context) error {

	log.Println("WeeklyCostReport job stared...")

	url := configuration.Get().Slack.WeeklyWebHookURL
	log.Println("weekly webhook rul:", url)

	return nil
}
