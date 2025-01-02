package configuration

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
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
