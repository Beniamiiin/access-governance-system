package commands

import (
	"access_governance_system/internal/db/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Command interface {
	CanHandle(command string) bool
	Handle(command, arguments string, user *models.User, bot *tgbotapi.BotAPI, chatID int64) []tgbotapi.Chattable
}
