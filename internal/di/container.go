package di

import (
	"context"
	zaploki "github.com/paul-milne/zap-loki"
	"go.uber.org/zap"
	"time"
)

func NewLogger(appName, environment, url string) *zap.SugaredLogger {
	if url == "" {
		return zap.Must(zap.NewProduction()).Sugar()
	}

	ctx := context.Background()
	lokiConfig := zaploki.Config{
		Url:          url,
		BatchMaxSize: 1000,
		BatchMaxWait: 10 * time.Second,
		Labels: map[string]string{
			"app":         appName,
			"environment": environment,
		},
	}
	return zap.Must(zaploki.New(ctx, lokiConfig).WithCreateLogger(zap.NewProductionConfig())).Sugar()
}
