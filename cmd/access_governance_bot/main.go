package main

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/tg_bot"
	"access_governance_system/internal/tg_bot/commands"
	"access_governance_system/internal/tg_bot/handlers"
	"context"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	go func() {
		logger.Info("setting up health check server")
		settingUpHealthCheckServer(logger)
	}()

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

func settingUpHealthCheckServer(logger *zap.SugaredLogger) {
	server := &http.Server{Addr: ":8080", Handler: http.HandlerFunc(handler)}

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
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

func handler(w http.ResponseWriter, r *http.Request) {}
