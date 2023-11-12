package main

import (
	"access_governance_system/configs"
	"access_governance_system/internal/di"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

var (
	config configs.DiscordAuthrozationBotConfig
	logger *zap.SugaredLogger
)

func main() {
	config, err := configs.LoadDiscordAuthrozationBotConfig()
	logger := di.NewLogger(config.Logger.AppName, config.App.Environment, config.Logger.URL)

	if err != nil {
		logger.Fatalw("failed to load config", "error", err)
	}
	logger.Info("config loaded")

	go func() {
		logger.Info("setting up health check server")
		settingUpHealthCheckServer(logger)
	}()

	logger.Info("starting bot")

	discord, err := discordgo.New("Bot " + config.AuthrozationBot.Token)
	if err != nil {
		logger.Fatalw("failed to create discord session", "error", err)
	}

	discord.AddHandler(authorization)

	discord.Identify.Intents = discordgo.IntentsGuildMessages

	err = discord.Open()
	if err != nil {
		logger.Fatalw("error opening connection", "error", err)
		return
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	discord.Close()
}

func authorization(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!authorize" {
		tgBotLink := fmt.Sprintf("https://t.me/S16AuthorizationBot?start=%s", m.Author.ID)
		message := fmt.Sprintf("Привет, для авторизации в сообществе %s перейди по ссылке %s", config.App.CommunityName, tgBotLink)

		channel, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			fmt.Println("failed to create author channel", "error", err)
		}

		_, err = s.ChannelMessageSend(channel.ID, message)
		if err != nil {
			fmt.Println("failed to send message", "error", err)
		}
	}
}

func settingUpHealthCheckServer(logger *zap.SugaredLogger) {
	mux := http.NewServeMux()
	mux.HandleFunc("/authorization-bot-discord/healthcheck", healthCheckHandler)

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
