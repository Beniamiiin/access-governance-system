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

const approvedProposalsCommandName = "approved_proposals"

type approvedProposalsCommand struct {
	proposalRepository repositories.ProposalRepository
	logger             *zap.SugaredLogger
}

func NewApprovedProposalsCommand(proposalRepository repositories.ProposalRepository, logger *zap.SugaredLogger) Command {
	return &approvedProposalsCommand{
		proposalRepository: proposalRepository,
		logger:             logger,
	}
}

func (c *approvedProposalsCommand) CanHandle(command string) bool {
	return command == approvedProposalsCommandName
}

func (c *approvedProposalsCommand) Start(text string, user *models.User, chatID int64) tgbotapi.Chattable {
	proposals, err := c.proposalRepository.GetManyByStatus(models.ProposalStatusApproved, models.ProposalStatusRejected)
	if err != nil {
		c.logger.Errorw("failed to get proposals", "error", err)
		return tgbot.DefaultErrorMessage(chatID)
	}

	message := ""

	if len(proposals) == 0 {
		message = "Нет одобренных предложений"
	} else {
		for _, proposal := range proposals {
			if user.Role == models.UserRoleSeeder {
				message += fmt.Sprintf("Тип: %s\n", proposal.NomineeRole.String())
			}

			message += fmt.Sprintf("Никнейм: %s\n", proposal.NomineeTelegramNickname)
			message += fmt.Sprintf("Дата начала: %s\n", internal.Format(proposal.CreatedAt))
			message += fmt.Sprintf("Дата окончания: %s\n", internal.Format(proposal.FinishedAt))
			message += fmt.Sprintf("Результат: %s\n", proposal.Status.String())
			message += fmt.Sprintln()
		}
	}

	return tgbotapi.NewMessage(chatID, message)
}
