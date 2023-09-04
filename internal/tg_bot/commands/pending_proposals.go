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

	var message tgbotapi.MessageConfig
	var messageText string

	if len(proposals) == 0 {
		messageText = "Нет предложений на рассмотрении"
		message = tgbotapi.NewMessage(chatID, messageText)
	} else {
		for _, proposal := range proposals {
			if user.Role == models.UserRoleSeeder {
				messageText += fmt.Sprintf("Тип: %s\n", proposal.NomineeRole.String())
			}

			messageText += fmt.Sprintf("Участник: %s (@%s)\n", proposal.NomineeName, proposal.NomineeTelegramNickname)

			if user.Role == models.UserRoleSeeder {
				messageText += fmt.Sprintf("Комментарий: %s\n", proposal.Comment)
			}

			messageText += fmt.Sprintf("Дата начала: %s\n", internal.Format(proposal.CreatedAt))
			messageText += fmt.Sprintf("Дата окончания: %s\n", internal.Format(proposal.FinishedAt))
			messageText += fmt.Sprintln()

			if user.Role == models.UserRoleSeeder {
				pollChatID := strings.TrimPrefix(strconv.Itoa(proposal.Poll.ChatID), "-100")
				messageText += fmt.Sprintf("Голосование можно найти [тут](https://t.me/c/%s/%d)\n", pollChatID, proposal.Poll.PollMessageID)
				messageText += fmt.Sprintf("Обсуждение можно найти [тут](https://t.me/c/%s/%d)\n", pollChatID, proposal.Poll.DiscussionMessageID)
			}

			message = tgbotapi.NewMessage(chatID, messageText)
			message.ParseMode = tgbotapi.ModeMarkdownV2
		}
	}

	return message
}
