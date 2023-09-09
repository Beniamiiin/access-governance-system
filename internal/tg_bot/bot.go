package tgbot

import (
	"access_governance_system/configs"
	"access_governance_system/internal/tg_bot/handlers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type bot struct {
	handler handlers.CommandHandler
}

type Bot interface {
	Start(config configs.AccessGovernanceBotConfig, logger *zap.SugaredLogger)
}

func NewBot(handler handlers.CommandHandler) Bot {
	return &bot{handler: handler}
}

func (b *bot) Start(config configs.AccessGovernanceBotConfig, logger *zap.SugaredLogger) {
	logger.Info("creating bot")
	bot, updates, err := b.createBot(config)
	if err != nil {
		logger.Fatalf("failed to create bot: %v", err)
	}
	logger.Info("bot created")

	for update := range updates {
		for _, message := range b.handler.Handle(update) {
			if _, err := bot.Send(message); err != nil {
				logger.Errorf("failed to send message: %v", err)
			}
		}
	}
}

func (b *bot) createBot(config configs.AccessGovernanceBotConfig) (*tgbotapi.BotAPI, tgbotapi.UpdatesChannel, error) {
	bot, err := tgbotapi.NewBotAPI(config.AccessGovernanceBot.Token)
	if err != nil {
		return nil, nil, err
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return bot, bot.GetUpdatesChan(u), nil
}
