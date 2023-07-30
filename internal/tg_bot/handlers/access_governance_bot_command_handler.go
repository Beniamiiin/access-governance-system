package handlers

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/tg_bot/commands"
	"errors"
	"fmt"
	"github.com/go-pg/pg/v10"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type accessGovernanceBotCommandHandler struct {
	appConfig      configs.App
	userRepository repositories.UserRepository
	logger         *zap.SugaredLogger
}

func NewAccessGovernanceBotCommandHandler(appConfig configs.App, userRepository repositories.UserRepository, logger *zap.SugaredLogger) CommandHandler {
	return &accessGovernanceBotCommandHandler{
		appConfig:      appConfig,
		userRepository: userRepository,
		logger:         logger,
	}
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

		isItSeeder := false

		if errors.Is(err, pg.ErrNoRows) {
			for _, seeder := range h.appConfig.InitialSeeders {
				if message.From.UserName == seeder {
					user = &models.User{
						TelegramID: message.From.ID,
						Role:       models.UserRoleSeeder,
					}

					user, err = h.userRepository.Create(user)
					if err != nil {
						h.logger.Errorw("failed to create user", "error", err)
						return nil
					}

					isItSeeder = true
					break
				}
			}
		}

		if !isItSeeder {
			return tgbotapi.NewMessage(chatID, fmt.Sprintf("Привет! К сожалению, ты не участник сообщества %s.", h.appConfig.CommunityName))
		}
	}

	if message.IsCommand() {
		command := message.Command()

		for _, handler := range commands {
			if handler.CanHandle(command) {
				user.TelegramState.LastCommand = command
				user, err = h.userRepository.Update(user)
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
