package handlers

import (
	"access_governance_system/internal/tg_bot/commands"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandHandler interface {
	Handle(commands []commands.Command, message *tgbotapi.Message) tgbotapi.Chattable
}
