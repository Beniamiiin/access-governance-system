package di

import (
	"access_governance_system/configs"
	"context"
	zaploki "github.com/paul-milne/zap-loki"
	"go.uber.org/zap"
	"time"
)

func NewLogger(config configs.Logger) *zap.SugaredLogger {
	if config.URL == "" {
		return zap.Must(zap.NewProduction()).Sugar()
	}

	ctx := context.Background()
	lokiConfig := zaploki.Config{
		Url:          config.URL,
		BatchMaxSize: 1000,
		BatchMaxWait: 10 * time.Second,
		Labels:       map[string]string{"app": config.AppName},
	}
	return zap.Must(zaploki.New(ctx, lokiConfig).WithCreateLogger(zap.NewProductionConfig())).Sugar()
}
