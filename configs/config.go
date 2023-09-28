package configs

import (
	"fmt"

	"github.com/caarlos0/env/v6"
)

type AccessGovernanceBotConfig struct {
	App                 App
	DB                  DB
	Logger              Logger
	AccessGovernanceBot AccessGovernanceBot
	VoteBot             VoteBot
	VoteAPI             VoteAPI
}

func LoadAccessGovernanceBotConfig() (AccessGovernanceBotConfig, error) {
	var config AccessGovernanceBotConfig

	if err := env.Parse(&config); err != nil {
		return AccessGovernanceBotConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

	config.Logger.AppName = "access-governance-bot"

	return config, nil
}

type ProposalStateServiceConfig struct {
	App                 App
	DB                  DB
	Logger              Logger
	AccessGovernanceBot AccessGovernanceBot
	VoteAPI             VoteAPI

	Quorum               float64 `env:"QUORUM"`                   // 30% initial parameter for quorum
	MinYesPercentage     float64 `env:"MIN_YES_PERCENTAGE"`       // Minimum 10% of votes should be "Yes"
	YesVotesToOvercomeNo float64 `env:"YES_VOTES_TO_OVERCOME_NO"` // 50% "yes" votes to overcome one "No vote"
}

func LoadProposalStateServiceConfig() (ProposalStateServiceConfig, error) {
	var config ProposalStateServiceConfig

	if err := env.Parse(&config); err != nil {
		return ProposalStateServiceConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

	config.Logger.AppName = "proposal-state-service"

	return config, nil
}
