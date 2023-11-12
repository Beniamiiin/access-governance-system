package agbcommands

import (
	"access_governance_system/internal"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/tg_bot/commands"
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

func NewApprovedProposalsCommand(proposalRepository repositories.ProposalRepository, logger *zap.SugaredLogger) commands.Command {
	return &approvedProposalsCommand{
		proposalRepository: proposalRepository,
		logger:             logger,
	}
}

func (c *approvedProposalsCommand) CanHandle(command string) bool {
	return command == approvedProposalsCommandName
}

func (c *approvedProposalsCommand) Handle(text, arguments string, user *models.User, chatID int64) []tgbotapi.Chattable {
	proposals, err := c.proposalRepository.GetManyByStatus(models.ProposalStatusApproved, models.ProposalStatusRejected)
	if err != nil {
		c.logger.Errorw("failed to get proposals", "error", err)
		return []tgbotapi.Chattable{tgbot.DefaultErrorMessage(chatID)}
	}

	messages := make([]tgbotapi.Chattable, 0, len(proposals))

	for _, proposal := range proposals {
		var messageText string

		if user.Role == models.UserRoleSeeder {
			messageText += fmt.Sprintf("Тип: %s\n", proposal.NomineeRole.String())
		}

		messageText += fmt.Sprintf("Участник: %s (@%s)\n", proposal.NomineeName, proposal.NomineeTelegramNickname)
		messageText += fmt.Sprintf("Дата начала: %s\n", internal.Format(proposal.CreatedAt))
		messageText += fmt.Sprintf("Дата окончания: %s\n", internal.Format(proposal.FinishedAt))
		messageText += fmt.Sprintf("Результат: %s\n", proposal.Status.String())

		message := tgbotapi.NewMessage(chatID, messageText)

		if user.Role == models.UserRoleSeeder {
			pollChatID := strings.TrimPrefix(strconv.Itoa(proposal.Poll.ChatID), "-100")

			message.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonURL("Посмотреть обсуждение", fmt.Sprintf("https://t.me/c/%s/%d", pollChatID, proposal.Poll.DiscussionMessageID)),
				),
			)
		}

		messages = append(messages, message)
	}

	if len(messages) == 0 {
		return []tgbotapi.Chattable{tgbotapi.NewMessage(chatID, "Нет одобренных предложений")}
	}

	return messages
}
