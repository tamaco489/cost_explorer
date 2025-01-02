package configuration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func loadSecrets(ctx context.Context, cfg Config) error {

	secretIDList := cfg.newSecretIDList()
	result, err := cfg.getFromSecretsManager(ctx, secretIDList)
	if err != nil {
		return err
	}

	for _, secret := range result.SecretValues {
		switch *secret.Name {
		case cfg.genSecretID(slackConfig.String()):
			if err := parseAndSetSlackConfig(secret.SecretString); err != nil {
				return err
			}

		case cfg.genSecretID(exchangeRatesAppID.String()):
			globalConfig.ExchangeRates.AppID = *secret.SecretString

		default:
			slog.WarnContext(ctx, "not found secret name",
				slog.String("env", cfg.Env),
				slog.String("secret name", *secret.Name),
			)
		}
	}

	return nil
}

type secretName string

const (
	exchangeRatesAppID secretName = "exchange-rates/app-id"
	slackConfig        secretName = "slack/config"
)

func (sn secretName) String() string {
	return string(sn)
}

func (cfg Config) newSecretIDList() []string {
	secretIDList := []string{
		cfg.genSecretID(exchangeRatesAppID.String()),
		cfg.genSecretID(slackConfig.String()),
	}

	return secretIDList
}

func (cfg Config) genSecretID(secretName string) string {
	return fmt.Sprintf("%s/%s/%s", cfg.ServiceName, cfg.Env, secretName)
}

func (cfg Config) getFromSecretsManager(ctx context.Context, secretIDList []string) (*secretsmanager.BatchGetSecretValueOutput, error) {

	svc := secretsmanager.NewFromConfig(cfg.AWSConfig)
	input := &secretsmanager.BatchGetSecretValueInput{
		SecretIdList: secretIDList,
	}

	result, err := svc.BatchGetSecretValue(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to batch-get secrets: %w", err)
	}

	return result, nil
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
