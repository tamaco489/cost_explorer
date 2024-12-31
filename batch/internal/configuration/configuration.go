package configuration

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Env         string `envconfig:"ENV" default:"dev"`
	ServiceName string `envconfig:"SERVICE_NAME" default:"cost-explorer"`
	AWSConfig   aws.Config
}

var globalConfig Config

func Get() Config { return globalConfig }

func Load(ctx context.Context) (Config, error) {

	envconfig.MustProcess("", &globalConfig)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := loadAWSConf(ctx); err != nil {
		return globalConfig, fmt.Errorf("failed to load aws config: %w", err)
	}

	return globalConfig, nil
}
