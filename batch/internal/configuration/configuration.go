package configuration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/kelseyhightower/envconfig"
)

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

type slackConfig struct {
	DailyWebHookURL  string `json:"daily_webhook_url"`
	WeeklyWebHookURL string `json:"weekly_webhook_url"`
}

var globalConfig Config

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

	if err := loadSlackConfig(ctx, globalConfig, globalConfig.Env); err != nil {
		return globalConfig, err
	}

	if err := loadExchangeRateAppID(ctx, globalConfig, globalConfig.Env); err != nil {
		return globalConfig, err
	}

	return globalConfig, nil
}

func loadSlackConfig(ctx context.Context, cfg Config, env string) error {
	secretName := fmt.Sprintf("cost-explorer/%s/slack/config", env)
	result, err := getFromSecretsManager(ctx, cfg.AWSConfig, secretName)
	if err != nil {
		return fmt.Errorf("failed to get slack config: %w", err)
	}

	var slackCfg slackConfig
	if err := json.Unmarshal([]byte(result), &slackCfg); err != nil {
		return fmt.Errorf("failed to parse slack config: %w", err)
	}

	globalConfig.Slack.DailyWebHookURL = slackCfg.DailyWebHookURL
	globalConfig.Slack.WeeklyWebHookURL = slackCfg.WeeklyWebHookURL

	return nil
}

func loadExchangeRateAppID(ctx context.Context, cfg Config, env string) error {
	secretName := fmt.Sprintf("cost-explorer/%s/exchange-rates/app-id", env)
	appID, err := getFromSecretsManager(ctx, cfg.AWSConfig, secretName)
	if err != nil {
		return fmt.Errorf("failed to get exchange rates app id: %w", err)
	}

	globalConfig.ExchangeRates.AppID = appID

	return nil
}
