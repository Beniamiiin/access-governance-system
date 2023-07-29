package commands

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	tgbot "access_governance_system/internal/tg_bot/extension"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"strings"
	"time"
)

const (
	createProposalCommandName = "create_proposal"

	waitingForTypeState     = "waiting_for_type"
	waitingForNicknameState = "waiting_for_nickname"
	waitingForReasonState   = "waiting_for_reason"
	waitingForConfirmState  = "waiting_for_confirm"
)

var (
	proposalTypeMember = models.UserRoleMember.CapitalizedString()
	proposalTypeSeeder = models.UserRoleSeeder.CapitalizedString()

	confirmYes = "Да"
	confirmNo  = "Нет, начать заново"
)

type createProposalCommand struct {
	appConfig          configs.App
	userRepository     repositories.UserRepository
	proposalRepository repositories.ProposalRepository
	logger             *zap.SugaredLogger
}

func NewCreateProposalCommand(appConfig configs.App, userRepository repositories.UserRepository, proposalRepository repositories.ProposalRepository, logger *zap.SugaredLogger) Command {
	return &createProposalCommand{
		appConfig:          appConfig,
		userRepository:     userRepository,
		proposalRepository: proposalRepository,
		logger:             logger,
	}
}

func (c *createProposalCommand) CanHandle(command string) bool {
	return command == createProposalCommandName
}

func (c *createProposalCommand) Start(text string, user *models.User, chatID int64) tgbotapi.Chattable {
	if text == createProposalCommandName {
		return c.handleCreateProposalCommand(user, chatID)
	}

	switch user.TelegramState.LastCommandState {
	case waitingForTypeState:
		return c.handleWaitingForTypeState(text, user, chatID)
	case waitingForNicknameState:
		return c.handleWaitingForNicknameState(text, user, chatID)
	case waitingForReasonState:
		return c.handleWaitingForReasonState(text, user, chatID)
	case waitingForConfirmState:
		return c.handleWaitingForConfirmState(text, user, chatID)
	default:
		c.logger.Errorf("user has unknown state: %s", user.TelegramState.LastCommandState)
		return nil
	}
}

func (c *createProposalCommand) handleCreateProposalCommand(user *models.User, chatID int64) tgbotapi.Chattable {
	user.TempProposal = models.Proposal{NominatorID: user.ID}

	switch user.Role {
	case models.UserRoleMember:
		return c.handleWaitingForTypeState(user.Role.CapitalizedString(), user, chatID)
	case models.UserRoleSeeder:
		message := tgbotapi.NewMessage(chatID, "Кого ты хочешь добавить? Выбери нужный тип участника.")
		message.ReplyMarkup = tgbotapi.NewOneTimeReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(proposalTypeMember),
				tgbotapi.NewKeyboardButton(proposalTypeSeeder),
			),
		)

		user.TelegramState.LastCommandState = waitingForTypeState
		_ = c.updateUser(user)

		return message
	}

	c.logger.Errorf("user has unknown role: %s", user.Role)
	return nil
}

func (c *createProposalCommand) handleWaitingForTypeState(proposalNomineeType string, user *models.User, chatID int64) tgbotapi.Chattable {
	message := tgbotapi.NewMessage(chatID, "Напиши, пожалуйста, никнейм пользователя в Telegram, которого ты хочешь предложить к добавлению в сообщество. Никнейм должен начинаться с @.")

	switch proposalNomineeType {
	case proposalTypeMember:
		user.TempProposal.NomineeRole = models.NomineeRoleMember
	case proposalTypeSeeder:
		user.TempProposal.NomineeRole = models.NomineeRoleSeeder
	default:
		c.logger.Warnf("user has unknown nominee type: %s", proposalNomineeType)
		return tgbotapi.NewMessage(chatID, fmt.Sprintf("Неизвестный тип участника: %s.", proposalNomineeType))
	}

	user.TelegramState.LastCommandState = waitingForNicknameState
	_ = c.updateUser(user)

	return message
}

func (c *createProposalCommand) handleWaitingForNicknameState(proposalNomineeNickname string, user *models.User, chatID int64) tgbotapi.Chattable {
	if !strings.HasPrefix(proposalNomineeNickname, "@") {
		c.logger.Warnf("user has invalid nominee nickname: %s", proposalNomineeNickname)
		return tgbotapi.NewMessage(chatID, "Никнейм должен начинаться с @.")
	}

	proposals, err := c.proposalRepository.GetManyByNomineeNickname(proposalNomineeNickname)
	if err != nil {
		c.logger.Errorf("failed to get proposals by nominee nickname: %s", err)
		return tgbot.ErrorMessage(chatID)
	} else if len(proposals) > 0 {
		lastProposal := proposals[len(proposals)-1]

		switch lastProposal.Status {
		case models.ProposalStatusCreated:
			c.logger.Warnf(
				"user tried to create proposal for nominee with existing created proposal: %s, %d, %s",
				proposalNomineeNickname,
				lastProposal.ID,
				lastProposal.CreatedAt,
			)
			return tgbotapi.NewMessage(chatID, "Предыдущее предложение на добавление это участника в сообщество ещё не рассмотрено.")
		case models.ProposalStatusApproved:
			c.logger.Warnf(
				"user tried to create proposal for nominee with existing approved proposal: %s, %d, %s",
				proposalNomineeNickname,
				lastProposal.ID,
				lastProposal.CreatedAt,
			)
			return tgbotapi.NewMessage(chatID, "Этот участник уже состоит в сообществе.")
		case models.ProposalStatusRejected:
			monthsAgo := time.Now().AddDate(0, -c.appConfig.RenominationPeriodMonths, 0)

			if !lastProposal.CreatedAt.Before(monthsAgo) {
				c.logger.Warnf(
					"user tried to create proposal for nominee with existing rejected proposal: %s, %d, %s",
					proposalNomineeNickname,
					lastProposal.ID,
					lastProposal.CreatedAt,
				)
				return tgbotapi.NewMessage(chatID, "Предыдущее предложение на добавление этого участника в сообщество было отклонено менее 3 месяцев назад. Участник может быть предложен к добавлению не чаще, чем раз в 3 месяца.")
			}
		}
	}

	message := tgbotapi.NewMessage(chatID, "Хорошо. Теперь напиши, почему ты считаешь, что этого участника стоит добавить в сообщество. Чем подробнее, тем лучше мы сможем понять твою точку зрения.")

	user.TempProposal.NomineeTelegramNickname = proposalNomineeNickname

	user.TelegramState.LastCommandState = waitingForReasonState
	_ = c.updateUser(user)

	return message
}

func (c *createProposalCommand) handleWaitingForReasonState(proposalDescription string, user *models.User, chatID int64) tgbotapi.Chattable {
	user.TempProposal.Comment = proposalDescription

	messageText := ""
	messageText += fmt.Sprintf("Тип: %s\n", user.TempProposal.NomineeRole)
	messageText += fmt.Sprintf("Никнейм: %s\n", user.TempProposal.NomineeTelegramNickname)
	messageText += fmt.Sprintf("Причина: %s\n", user.TempProposal.Comment)
	messageText += fmt.Sprintln()
	messageText += "Все правильно, отправляем предложение на голосование?"
	message := tgbotapi.NewMessage(chatID, messageText)
	message.ReplyMarkup = tgbotapi.NewOneTimeReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(confirmYes),
			tgbotapi.NewKeyboardButton(confirmNo),
		),
	)

	user.TelegramState.LastCommandState = waitingForConfirmState
	_ = c.updateUser(user)

	return message
}

func (c *createProposalCommand) handleWaitingForConfirmState(confirmationState string, user *models.User, chatID int64) tgbotapi.Chattable {
	if confirmationState == confirmNo {
		return c.Start(createProposalCommandName, user, chatID)
	}

	proposal, err := c.proposalRepository.Create(&user.TempProposal)
	if err != nil {
		c.logger.Errorw("failed to create proposal", "error", err)
		return tgbot.ErrorMessage(chatID)
	}

	proposal.FinishedAt = proposal.CreatedAt.AddDate(0, 0, c.appConfig.VotingDurationDays)
	if _, err = c.proposalRepository.Update(proposal); err != nil {
		c.logger.Errorw("failed to update proposal", "error", err)

		if err = c.proposalRepository.Delete(proposal); err != nil {
			c.logger.Errorw("failed to delete proposal", "error", err)
		}

		return tgbot.ErrorMessage(chatID)
	}

	c.logger.Info("proposal created")

	user.Proposals = append(user.Proposals, user.TempProposal)

	if err = c.updateUser(user); err != nil {
		c.logger.Errorw("failed to update user", "error", err)
		return tgbot.ErrorMessage(chatID)
	}
	c.logger.Info("user updated")

	message := tgbotapi.NewMessage(chatID, "Отлично, твое предложение отправлено на голосование.")
	message.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

	user.TempProposal = models.Proposal{}
	user.TelegramState = models.TelegramState{}
	_ = c.updateUser(user)

	return message
}

func (c *createProposalCommand) updateUser(user *models.User) error {
	_, err := c.userRepository.Update(user)
	if err != nil {
		c.logger.Errorw("failed to update user", "error", err)
	}
	return err
}
