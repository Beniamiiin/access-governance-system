package handlers

import (
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/tg_bot/commands"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type accessGovernanceBotCommandHandler struct {
	userRepository repositories.UserRepository
	logger         *zap.SugaredLogger
}

func NewAccessGovernanceBotCommandHandler(userRepository repositories.UserRepository, logger *zap.SugaredLogger) CommandHandler {
	return &accessGovernanceBotCommandHandler{userRepository: userRepository, logger: logger}
}

func (h *accessGovernanceBotCommandHandler) Handle(commands []commands.Command, message *tgbotapi.Message) tgbotapi.Chattable {
	h.logger.Info("received message")

	if message == nil {
		h.logger.Warn("received non-Message updates")
		return nil
	}

	chatID := message.Chat.ID

	user, err := h.userRepository.GetOne(message.From.ID)
	if user == nil || err != nil {
		h.logger.Errorw("failed to get user", "error", err)
		return tgbotapi.NewMessage(chatID, "Ты не зарегистрирован в системе")
	}

	if message.IsCommand() {
		command := message.Command()

		for _, handler := range commands {
			if handler.CanHandle(command) {
				user.TelegramState.LastCommand = command
				_, err := h.userRepository.Update(user)
				if err != nil {
					h.logger.Errorw("failed to update user", "error", err)
				}

				return handler.Start(command, user, chatID)
			}
		}

		h.logger.Errorf("received unknown command: %s", command)
		return nil
	} else if user.TelegramState.LastCommand != "" {
		command := user.TelegramState.LastCommand
		subCommand := message.Text

		for _, handler := range commands {
			if handler.CanHandle(command) {
				responseMessage := handler.Start(subCommand, user, chatID)
				if responseMessage == nil {
					h.logger.Errorf("failed to handle subcommand: %s", subCommand)
					break
				}
				return responseMessage
			}
		}

		h.logger.Errorf("received unknown subcommand: %s for command: %s", subCommand, command)
		return nil
	}

	h.logger.Error("received unknown message")
	return nil
}
