package utils

import (
	"context"
	"os"
)

func PrepareTestEnvironment(ctx context.Context) {
	if err := os.Setenv("ENV", "test"); err != nil {
		panic(err)
	}
}
