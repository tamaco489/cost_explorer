package configuration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/kelseyhightower/envconfig"
)

const (
	baseSecretName          = "cost-explorer"
	exchangeRatesSecretName = "exchange-rates/app-id"
	slackConfigSecretName   = "slack/config"
)

var secretIdList []string

func init() {
	env := globalConfig.Env
	secretIdList = []string{
		generateSecretName(env, exchangeRatesSecretName),
		generateSecretName(env, slackConfigSecretName),
	}
}

func generateSecretName(env, secretName string) string {
	secret := fmt.Sprintf("%s/%s/%s", baseSecretName, env, secretName)
	return secret
}

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

	if err := loadSecrets(ctx, globalConfig.AWSConfig, globalConfig.Env); err != nil {
		return globalConfig, err
	}

	return globalConfig, nil
}

func loadSecrets(ctx context.Context, awsConfig aws.Config, env string) error {
	result, err := getFromSecretsManager(ctx, awsConfig, secretIdList)
	if err != nil {
		return err
	}

	for _, secret := range result.SecretValues {
		switch *secret.Name {
		case generateSecretName(env, slackConfigSecretName):
			if err := parseAndSetSlackConfig(secret.SecretString); err != nil {
				return err
			}

		case generateSecretName(env, exchangeRatesSecretName):
			globalConfig.ExchangeRates.AppID = *secret.SecretString

		default:
			slog.WarnContext(ctx, "not found secret name",
				slog.String("env", env),
				slog.String("secret name", *secret.Name),
			)
		}
	}

	return nil
}

func parseAndSetSlackConfig(secretString *string) error {
	var slackConfig struct {
		DailyWebHookURL  string `json:"daily_webhook_url"`
		WeeklyWebHookURL string `json:"weekly_webhook_url"`
	}
	if err := json.Unmarshal([]byte(*secretString), &slackConfig); err != nil {
		return fmt.Errorf("failed to parse slack config: %w", err)
	}

	globalConfig.Slack.DailyWebHookURL = slackConfig.DailyWebHookURL
	globalConfig.Slack.WeeklyWebHookURL = slackConfig.WeeklyWebHookURL

	return nil
}
