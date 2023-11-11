package main

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/di"
	tgbot "access_governance_system/internal/tg_bot"
	"access_governance_system/internal/tg_bot/commands"
	abcommands "access_governance_system/internal/tg_bot/commands/authorization_bot"
	abhandlers "access_governance_system/internal/tg_bot/handlers/authorization_bot"
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
	config, err := configs.LoadAuthrozationBotConfig()
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

	tgbot.NewBot(
		abhandlers.NewAuthorizationBotCommandHandler(config.App, userRepository, logger,
			[]commands.Command{
				abcommands.NewStartCommand(config.DiscordAuthrozationBot, userRepository, logger),
			},
		),
	).Start(config.TelegramAuthrozationBot.Token, logger)
}

func settingUpHealthCheckServer(logger *zap.SugaredLogger) {
	mux := http.NewServeMux()
	mux.HandleFunc("/authorization-bot/healthcheck", healthCheckHandler)

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
