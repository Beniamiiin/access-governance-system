package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandHandler interface {
	Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) []tgbotapi.Chattable
}
