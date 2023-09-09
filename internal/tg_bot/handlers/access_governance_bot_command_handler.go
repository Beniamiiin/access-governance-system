package handlers

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/tg_bot/commands"
	tgbot "access_governance_system/internal/tg_bot/extension"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type accessGovernanceBotCommandHandler struct {
	appConfig      configs.App
	userRepository repositories.UserRepository
	logger         *zap.SugaredLogger

	commands []commands.Command
}

func NewAccessGovernanceBotCommandHandler(
	appConfig configs.App,
	userRepository repositories.UserRepository,
	logger *zap.SugaredLogger,
	commands []commands.Command,
) CommandHandler {
	return &accessGovernanceBotCommandHandler{
		appConfig:      appConfig,
		userRepository: userRepository,
		logger:         logger,
		commands:       commands,
	}
}

func (h *accessGovernanceBotCommandHandler) Handle(update tgbotapi.Update) []tgbotapi.Chattable {
	h.logger.Info("received message")

	message := update.Message
	callbackQuery := update.CallbackQuery

	if message == nil && callbackQuery == nil {
		h.logger.Warn("received unknown updates")
		return []tgbotapi.Chattable{}
	}

	var (
		chatID       int64
		telegramUser *tgbotapi.User
	)

	if message != nil {
		chatID = message.Chat.ID
		telegramUser = message.From
	} else if callbackQuery != nil {
		chatID = callbackQuery.Message.Chat.ID
		telegramUser = callbackQuery.From
	}

	user, errMessage := h.createUserIfNeeded(telegramUser.ID, telegramUser.UserName, chatID)
	if errMessage != nil {
		return []tgbotapi.Chattable{errMessage}
	}

	if message != nil {
		if message.IsCommand() {
			return h.tryToHandleCommand(message.Command(), h.commands, user, chatID)
		} else if user.TelegramState.LastCommand != "" {
			return h.tryToHandleSubCommand(user.TelegramState.LastCommand, message.Text, h.commands, user, chatID)
		}
	}

	if callbackQuery != nil {
		return h.tryToHandleQueryCallback(callbackQuery.Data, h.commands, user, chatID)
	}

	h.logger.Error("received unknown message")
	return []tgbotapi.Chattable{}
}

func (h *accessGovernanceBotCommandHandler) createUserIfNeeded(userID int64, userName string, chatID int64) (*models.User, tgbotapi.Chattable) {
	user, err := h.userRepository.GetOneByTelegramID(userID)
	if err != nil {
		h.logger.Errorw("failed to get user", "error", err)
	}

	if user == nil {
		isItSeeder := false

		for _, seederName := range h.appConfig.InitialSeeders {
			if userName == seederName {
				user = &models.User{
					TelegramID:       userID,
					TelegramNickname: userName,
					Role:             models.UserRoleSeeder,
				}

				user, err = h.userRepository.Create(user)
				if err != nil {
					h.logger.Errorw("failed to create user", "error", err)
					return nil, tgbot.DefaultErrorMessage(chatID)
				}

				isItSeeder = true
				break
			}
		}

		if !isItSeeder {
			return nil, tgbotapi.NewMessage(chatID, fmt.Sprintf("Привет! К сожалению, ты не участник сообщества %s.", h.appConfig.CommunityName))
		}
	}

	return user, nil
}

func (h *accessGovernanceBotCommandHandler) tryToHandleCommand(command string, commands []commands.Command, user *models.User, chatID int64) []tgbotapi.Chattable {
	for _, handler := range commands {
		if handler.CanHandle(command) {
			user.TelegramState.LastCommand = command

			user, err := h.userRepository.Update(user)
			if err != nil {
				h.logger.Errorw("failed to update user", "error", err)
			}

			return handler.Handle(command, user, chatID)
		}
	}

	h.logger.Errorf("received unknown command: %s", command)
	return []tgbotapi.Chattable{}
}

func (h *accessGovernanceBotCommandHandler) tryToHandleSubCommand(command, subCommand string, commands []commands.Command, user *models.User, chatID int64) []tgbotapi.Chattable {
	command = strings.Split(command, ":")[0]

	for _, handler := range commands {
		if handler.CanHandle(command) {
			responseMessage := handler.Handle(subCommand, user, chatID)
			if responseMessage == nil {
				h.logger.Errorf("failed to handle subcommand: %s", subCommand)
				break
			}

			return responseMessage
		}
	}

	h.logger.Errorf("received unknown subcommand: %s for command: %s", subCommand, command)
	return []tgbotapi.Chattable{}
}

func (h *accessGovernanceBotCommandHandler) tryToHandleQueryCallback(query string, commands []commands.Command, user *models.User, chatID int64) []tgbotapi.Chattable {
	parts := strings.Split(query, ":")
	if len(parts) == 0 {
		h.logger.Error("received empty query callback")
		return []tgbotapi.Chattable{}
	}

	command := parts[0]

	for _, handler := range commands {
		if handler.CanHandle(command) {
			user.TelegramState.LastCommand = query

			user, err := h.userRepository.Update(user)
			if err != nil {
				h.logger.Errorw("failed to update user", "error", err)
			}

			return handler.Handle(query, user, chatID)
		}
	}

	h.logger.Errorf("received unknown command: %s", command)
	return []tgbotapi.Chattable{}
}
