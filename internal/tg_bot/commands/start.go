package commands

import (
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const startCommandName = "start"

type startCommand struct {
	userRepository repositories.UserRepository
	logger         *zap.SugaredLogger
}

func NewStartCommand(userRepository repositories.UserRepository, logger *zap.SugaredLogger) Command {
	return &startCommand{
		userRepository: userRepository,
		logger:         logger,
	}
}

func (c *startCommand) CanHandle(command string) bool {
	return command == startCommandName
}

func (c *startCommand) Start(text string, user *models.User, chatID int64) tgbotapi.Chattable {
	return tgbotapi.NewMessage(chatID, "Привет!")
}
