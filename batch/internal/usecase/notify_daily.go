package usecase

import (
	"context"
	"log"
)

func (j *Job) NotifyDaily(ctx context.Context) error {

	log.Println("NotifyDaily job started...")

	return nil
}
