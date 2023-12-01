package abhandlers

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/tg_bot/commands"
	"access_governance_system/internal/tg_bot/extension"
	"access_governance_system/internal/tg_bot/handlers"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type authorizationBotCommandHandler struct {
	appConfig      configs.App
	userRepository repositories.UserRepository
	logger         *zap.SugaredLogger

	commands []commands.Command
}

func NewAuthorizationBotCommandHandler(
	appConfig configs.App,
	userRepository repositories.UserRepository,
	logger *zap.SugaredLogger,
	commands []commands.Command,
) handlers.CommandHandler {
	return &authorizationBotCommandHandler{
		appConfig:      appConfig,
		userRepository: userRepository,
		logger:         logger,
		commands:       commands,
	}
}

func (h *authorizationBotCommandHandler) Handle(bot *tgbotapi.BotAPI, update tgbotapi.Update) []tgbotapi.Chattable {
	h.logger.Info("received message")

	message := update.Message

	if message == nil {
		h.logger.Warn("received unknown updates")
		return []tgbotapi.Chattable{}
	}

	chatID := message.Chat.ID

	user, err := h.userRepository.GetOneByTelegramID(message.From.ID)
	if err != nil {
		h.logger.Errorw("failed to get user", "error", err)
		return []tgbotapi.Chattable{extension.DefaultErrorMessage(chatID)}
	} else if user == nil {
		h.logger.Warnw("failed to get user", "error", err)
		return []tgbotapi.Chattable{
			tgbotapi.NewMessage(chatID, fmt.Sprintf("Привет! К сожалению, ты не участник сообщества %s.", h.appConfig.CommunityName)),
		}
	}

	if message != nil {
		h.logger.Infow("received message", "message", message)

		if message.IsCommand() {
			h.logger.Infow("received command", "command", message.Command())
			return h.tryToHandleCommand(message, h.commands, user, chatID)
		}
	}

	h.logger.Warn("received unknown message")
	return []tgbotapi.Chattable{}
}

func (h *authorizationBotCommandHandler) tryToHandleCommand(message *tgbotapi.Message, commands []commands.Command, user *models.User, chatID int64) []tgbotapi.Chattable {
	command := message.Command()
	arguments := message.CommandArguments()

	for _, handler := range commands {
		if handler.CanHandle(command) {
			return handler.Handle(command, arguments, user, chatID)
		}
	}

	h.logger.Warnw("received unknown command", "command", command)
	return []tgbotapi.Chattable{}
}
