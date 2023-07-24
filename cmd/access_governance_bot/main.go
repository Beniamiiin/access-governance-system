package main

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/tg_bot"
	"access_governance_system/internal/tg_bot/commands"
	"access_governance_system/internal/tg_bot/handlers"
	"go.uber.org/zap"
)

func main() {
	logger := zap.Must(zap.NewProduction()).Sugar()

	logger.Info("loading config")
	config, err := configs.LoadAccessGovernanceBotConfig()
	if err != nil {
		logger.Fatalw("failed to load config", "error", err)
	}
	logger.Info("config loaded")

	logger.Info("starting db")
	database, err := db.StartDB(config.DB, logger)
	if err != nil {
		logger.Fatalw("failed to start db", "error", err)
	}
	logger.Info("db started")

	logger.Info("starting bot")
	userRepository := repositories.NewUserRepository(database)
	proposalRepository := repositories.NewProposalRepository(database)

	tg_bot.NewBot(
		[]commands.Command{
			commands.NewStartCommand(userRepository, logger),
			commands.NewApprovedProposalsCommand(proposalRepository, logger),
			commands.NewCreateProposalCommand(userRepository, proposalRepository, logger),
			commands.NewPendingProposalsCommand(proposalRepository, logger),
		},
		handlers.NewAccessGovernanceBotCommandHandler(userRepository, logger),
	).Start(config, logger)
}
