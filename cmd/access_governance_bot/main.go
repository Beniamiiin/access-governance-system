package main

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/di"
	"access_governance_system/internal/services"
	tgbot "access_governance_system/internal/tg_bot"
	"access_governance_system/internal/tg_bot/commands"
	agbcommands "access_governance_system/internal/tg_bot/commands/access_governance_bot"
	agbhandlers "access_governance_system/internal/tg_bot/handlers/access_governance_bot"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	config, err := configs.LoadAccessGovernanceBotConfig()
	logger := di.NewLogger(config.Logger.AppName, config.App.Environment, config.Logger.URL)

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

	go func() {
		logger.Info("setting up health check server")
		settingUpHealthCheckServer(logger)
	}()

	logger.Info("starting bot")
	userRepository := repositories.NewUserRepository(database)
	proposalRepository := repositories.NewProposalRepository(database)
	voteService := services.NewVoteService(config.VoteAPI.URL)

	tgbot.NewBot(
		agbhandlers.NewAccessGovernanceBotCommandHandler(config, userRepository, proposalRepository, logger,
			[]commands.Command{
				agbcommands.NewStartCommand(config, userRepository, logger),
				agbcommands.NewCancelProposalCommand(config.App, userRepository, logger),
				agbcommands.NewApprovedProposalsCommand(proposalRepository, logger),
				agbcommands.NewCreateProposalCommand(config, userRepository, proposalRepository, voteService, logger),
				agbcommands.NewPendingProposalsCommand(userRepository, proposalRepository, logger),
				agbcommands.NewAddCommentCommand(userRepository, proposalRepository, config.VoteBot, logger),
			},
		),
	).Start(config.AccessGovernanceBot.Token, logger)
}

func settingUpHealthCheckServer(logger *zap.SugaredLogger) {
	mux := http.NewServeMux()
	mux.HandleFunc("/access-governance-bot/healthcheck", healthCheckHandler)

	server := &http.Server{Addr: ":8080", Handler: mux}

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logger.Errorw("failed to start http server", "error", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		logger.Errorw("failed to shutdown http server", "error", err)
		return
	}

	logger.Info("shutting down")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("I'm alive"))
}
