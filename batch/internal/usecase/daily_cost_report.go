package usecase

import (
	"context"
	"log"
)

func (j *Job) DailyCostReport(ctx context.Context) error {

	log.Println("DailyCostReport job started...")

	return nil
}
