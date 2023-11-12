package commands

import (
	"access_governance_system/internal/db/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Command interface {
	CanHandle(command string) bool
	Handle(text, arguments string, user *models.User, chatID int64) []tgbotapi.Chattable
}
