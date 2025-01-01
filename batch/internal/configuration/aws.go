package configuration

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func loadAWSConf(ctx context.Context) error {
	const awsRegion = "ap-northeast-1"
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		return fmt.Errorf("failed to load aws config: %w", err)
	}
	globalConfig.AWSConfig = cfg

	return nil
}

var awsSecretCache = make(map[string]string)

func getFromSecretsManager(ctx context.Context, awsConfig aws.Config, secretName string) (string, error) {
	c, exists := awsSecretCache[secretName]
	if exists {
		return c, nil
	}

	svc := secretsmanager.NewFromConfig(awsConfig)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := svc.GetSecretValue(ctx, input)
	if err != nil {
		return "", err
	}

	awsSecretCache[secretName] = *result.SecretString

	return *result.SecretString, nil
}
