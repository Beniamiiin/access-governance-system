package commands

import (
	"access_governance_system/internal"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	tgbot "access_governance_system/internal/tg_bot/extension"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const pendingProposalsCommandName = "pending_proposals"

type pendingProposalsCommand struct {
	proposalRepository repositories.ProposalRepository
	logger             *zap.SugaredLogger
}

func NewPendingProposalsCommand(proposalRepository repositories.ProposalRepository, logger *zap.SugaredLogger) Command {
	return &pendingProposalsCommand{
		proposalRepository: proposalRepository,
		logger:             logger,
	}
}

func (c *pendingProposalsCommand) CanHandle(command string) bool {
	return command == pendingProposalsCommandName
}

func (c *pendingProposalsCommand) Start(text string, user *models.User, chatID int64) tgbotapi.Chattable {
	proposals, err := c.proposalRepository.GetManyByStatus(models.ProposalStatusCreated)
	if err != nil {
		c.logger.Errorw("failed to get proposals", "error", err)
		return tgbot.DefaultErrorMessage(chatID)
	}

	var message string

	if len(proposals) == 0 {
		message = "Нет предложений на рассмотрении"
	} else {
		for _, proposal := range proposals {
			if user.Role == models.UserRoleSeeder {
				message += fmt.Sprintf("Тип: %s\n", proposal.NomineeRole.String())
			}

			message += fmt.Sprintf("Участник: %s (@%s)\n", proposal.NomineeName, proposal.NomineeTelegramNickname)

			if user.Role == models.UserRoleSeeder {
				message += fmt.Sprintf("Комментарий: %s\n", proposal.Comment)
			}

			message += fmt.Sprintf("Дата начала: %s\n", internal.Format(proposal.CreatedAt))
			message += fmt.Sprintf("Дата окончания: %s\n", internal.Format(proposal.FinishedAt))
			message += fmt.Sprintln()
		}
	}

	return tgbotapi.NewMessage(chatID, message)
}
