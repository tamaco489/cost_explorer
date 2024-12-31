package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
	"github.com/tamaco489/cost_explorer/batch/internal/handler"
	"github.com/tamaco489/cost_explorer/batch/internal/usecase"
)

func main() {
	ctx := context.Background()
	cfg, err := configuration.Load(ctx)
	if err != nil {
		panic(err)
	}

	job, err := usecase.NewJob(cfg)
	if err != nil {
		panic(err)
	}

	lambda.Start(handler.JobHandler(*job))
}
