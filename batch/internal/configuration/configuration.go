package configuration

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/kelseyhightower/envconfig"
)

var globalConfig Config

type Config struct {
	Env         string `envconfig:"ENV" default:"dev"`
	ServiceName string `envconfig:"SERVICE_NAME" default:"cost-explorer"`
	Slack       struct {
		DailyWebHookURL  string
		WeeklyWebHookURL string
	}
	ExchangeRates struct {
		AppID string
	}
	Logging   string `envconfig:"LOGGING" default:"off"`
	AWSConfig aws.Config
}

func Get() Config {
	return globalConfig
}

func Load(ctx context.Context) (Config, error) {
	envconfig.MustProcess("", &globalConfig)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := loadAWSConf(ctx); err != nil {
		return globalConfig, err
	}
	if err := loadSecrets(ctx, globalConfig); err != nil {
		return globalConfig, err
	}

	return globalConfig, nil
}
