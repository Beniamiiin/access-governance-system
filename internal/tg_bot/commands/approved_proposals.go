package commands

import (
	"access_governance_system/internal"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	tgbot "access_governance_system/internal/tg_bot/extension"
	"fmt"
	"strconv"
	"strings"

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

	var message tgbotapi.MessageConfig
	var messageText string

	if len(proposals) == 0 {
		messageText = "Нет одобренных предложений"
		message = tgbotapi.NewMessage(chatID, messageText)
	} else {
		for _, proposal := range proposals {
			if user.Role == models.UserRoleSeeder {
				messageText += fmt.Sprintf("Тип: %s\n", proposal.NomineeRole.String())
			}

			messageText += fmt.Sprintf("Участник: %s (@%s)\n", proposal.NomineeName, proposal.NomineeTelegramNickname)
			messageText += fmt.Sprintf("Дата начала: %s\n", internal.Format(proposal.CreatedAt))
			messageText += fmt.Sprintf("Дата окончания: %s\n", internal.Format(proposal.FinishedAt))
			messageText += fmt.Sprintf("Результат: %s\n", proposal.Status.String())
			if user.Role == models.UserRoleSeeder {
				pollChatID := strings.TrimPrefix(strconv.Itoa(proposal.Poll.ChatID), "-100")
				messageText += fmt.Sprintf("Обсуждение можно найти [тут](https://t.me/c/%s/%d)\n", pollChatID, proposal.Poll.DiscussionMessageID)
			}

			messageText += fmt.Sprintln()

			parseMode := tgbotapi.ModeMarkdownV2
			messageText = tgbotapi.EscapeText(parseMode, messageText)
			message = tgbotapi.NewMessage(chatID, messageText)
			message.ParseMode = parseMode
		}
	}

	return message
}
