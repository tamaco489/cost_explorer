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

func getFromSecretsManager(ctx context.Context, awsConfig aws.Config, secretIdList []string) (*secretsmanager.BatchGetSecretValueOutput, error) {

	svc := secretsmanager.NewFromConfig(awsConfig)
	input := &secretsmanager.BatchGetSecretValueInput{
		SecretIdList: secretIdList,
	}

	result, err := svc.BatchGetSecretValue(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to batch-get secrets: %w", err)
	}

	return result, nil
}
