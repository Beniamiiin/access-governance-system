package configs

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

type AccessGovernanceBotConfig struct {
	App                 App
	DB                  DB
	AccessGovernanceBot Bot
	VoteBot             Bot
	VoteAPI             VoteAPI

	DiscordInviteLink string `env:"DISCORD_INVITE_LINK"`
}

func LoadAccessGovernanceBotConfig() (AccessGovernanceBotConfig, error) {
	var config AccessGovernanceBotConfig

	if err := env.Parse(&config); err != nil {
		return AccessGovernanceBotConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

	config.AccessGovernanceBot.Token = os.Getenv("TELEGRAM_ACCESS_GOVERNANCE_BOT_TOKEN")
	config.VoteBot.Token = os.Getenv("TELEGRAM_VOTE_BOT_TOKEN")

	return config, nil
}

type ProposalStateServiceConfig struct {
	App                 App
	DB                  DB
	Logger              Logger
	AccessGovernanceBot Bot
	VoteAPI             VoteAPI

	Quorum                  float64 `env:"QUORUM"`                     // 30% initial parameter for quorum
	MaxRequiredSeedersCount float64 `env:"MAX_REQUIRED_SEEDERS_COUNT"` // But not more than 10 votes
	MinYesVotesPercentage   float64 `env:"MIN_YES_VOTES_PERCENTAGE"`   // Minimum 10% of votes should be "Yes"
	MinRequiredYesVotes     float64 `env:"MIN_REQUIRED_YES_VOTES"`     // But not less than 3
	YesVotesToOvercomeNo    float64 `env:"YES_VOTES_TO_OVERCOME_NO"`   // 50% "yes" votes to overcome one "No vote"
}

func LoadProposalStateServiceConfig() (ProposalStateServiceConfig, error) {
	var config ProposalStateServiceConfig

	if err := env.Parse(&config); err != nil {
		return ProposalStateServiceConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

	config.AccessGovernanceBot.Token = os.Getenv("TELEGRAM_ACCESS_GOVERNANCE_BOT_TOKEN")
	config.Logger.AppName = "proposal-state-service"

	return config, nil
}

type TelegramAuthrozationBotConfig struct {
	App                     App
	DB                      DB
	Logger                  Logger
	DiscordAuthrozationBot  Discord
	TelegramAuthrozationBot Bot
}

func LoadTelegramAuthrozationBotConfig() (TelegramAuthrozationBotConfig, error) {
	var config TelegramAuthrozationBotConfig

	if err := env.Parse(&config); err != nil {
		return TelegramAuthrozationBotConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

	config.TelegramAuthrozationBot.Token = os.Getenv("TELEGRAM_AUTHORIZATION_BOT_TOKEN")
	config.Logger.AppName = "authorization-bot-telegram"

	return config, nil
}

type DiscordAuthrozationBotConfig struct {
	App             App
	Logger          Logger
	AuthrozationBot Discord
}

func LoadDiscordAuthrozationBotConfig() (DiscordAuthrozationBotConfig, error) {
	var config DiscordAuthrozationBotConfig

	if err := env.Parse(&config); err != nil {
		return DiscordAuthrozationBotConfig{}, fmt.Errorf("failed to parse config: %w", err)
	}

	config.Logger.AppName = "authorization-bot-discord"

	return config, nil
}
