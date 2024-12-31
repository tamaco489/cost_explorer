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
		WebHookURL string
	}
	AWSConfig aws.Config
}

type SlackConfig struct {
	WebHookURL string `json:"webhook_url"`
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
		return globalConfig, fmt.Errorf("failed to load aws config: %w", err)
	}

	if err := loadSlackConfig(ctx, globalConfig, globalConfig.Env); err != nil {
		return globalConfig, fmt.Errorf("failed to load slack config: %w", err)
	}

	return globalConfig, nil
}

func loadSlackConfig(ctx context.Context, cfg Config, env string) error {
	secretName := fmt.Sprintf("cost-explorer/%s/slack", env)
	result, err := getFromSecretsManager(ctx, cfg.AWSConfig, secretName)
	if err != nil {
		return fmt.Errorf("failed to get slack config: %w", err)
	}

	var slackCfg SlackConfig
	if err := json.Unmarshal([]byte(result), &slackCfg); err != nil {
		return fmt.Errorf("failed to parse slack config: %w", err)
	}

	globalConfig.Slack.WebHookURL = slackCfg.WebHookURL

	return nil
}
