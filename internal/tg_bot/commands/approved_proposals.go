package commands

import (
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
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
	proposals, err := c.proposalRepository.GetMany(models.ProposalStatusApproved)
	if err != nil {
		c.logger.Errorw("failed to get proposals", "error", err)
		return tgbotapi.NewMessage(chatID, "Произошла ошибка, повторите попытку позже")
	}

	var message string

	if len(proposals) == 0 {
		message = "Нет одобренных предложений"
	} else {
		for _, proposal := range proposals {
			message += fmt.Sprintf("Proposal ID: %d\n", proposal.ID)
		}
	}

	return tgbotapi.NewMessage(chatID, message)
}
