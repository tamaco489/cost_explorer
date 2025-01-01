package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/tamaco489/cost_explorer/batch/internal/configuration"
	"github.com/tamaco489/cost_explorer/batch/internal/handler"
	"github.com/tamaco489/cost_explorer/batch/internal/usecase"
)

func main() {
	// json形式でログ出力するためのハンドラーを生成
	h := slog.NewJSONHandler(
		os.Stdout,                             // Lambdaでは標準出力で書き込んだログはCloudWatchに送信されるため、Lambda環境下で適切にログを記録するための設定
		&slog.HandlerOptions{AddSource: true}, // ログメッセージに発生元の情報（ファイル名や行番号など）を自動的に追加する設定
	)

	slog.SetDefault(slog.New(h))

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
