package configs

import (
	"fmt"
	"github.com/caarlos0/env/v6"
)

type AccessGovernanceBotConfig struct {
	App App
	Bot Bot
	DB  DB
}

func LoadAccessGovernanceBotConfig() (AccessGovernanceBotConfig, error) {
	var config AccessGovernanceBotConfig

	if err := env.Parse(&config); err != nil {
		return AccessGovernanceBotConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}
