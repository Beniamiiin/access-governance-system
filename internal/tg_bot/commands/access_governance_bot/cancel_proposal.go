package agbcommands

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/tg_bot/commands"
	tgbot "access_governance_system/internal/tg_bot/extension"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const cancelProposalCommandName = "cancel_proposal"

type cancelProposalCommand struct {
	appConfig      configs.App
	userRepository repositories.UserRepository
	logger         *zap.SugaredLogger
}

func NewCancelProposalCommand(appConfig configs.App, userRepository repositories.UserRepository, logger *zap.SugaredLogger) commands.Command {
	return &cancelProposalCommand{
		appConfig:      appConfig,
		userRepository: userRepository,
		logger:         logger,
	}
}

func (c *cancelProposalCommand) CanHandle(command string) bool {
	return command == cancelProposalCommandName
}

func (c *cancelProposalCommand) Handle(text, arguments string, user *models.User, chatID int64) []tgbotapi.Chattable {
	user.TempProposal = models.Proposal{}
	user.TelegramState = models.TelegramState{}
	_, err := c.userRepository.Update(user)
	if err != nil {
		c.logger.Errorw("failed to update user", "error", err)
		return []tgbotapi.Chattable{tgbot.DefaultErrorMessage(chatID)}
	}

	message := tgbotapi.NewMessage(chatID, "Предыдущее незавершенное предложение удалено. Выберите команду /create_proposal для создания нового.")
	return []tgbotapi.Chattable{message}
}
