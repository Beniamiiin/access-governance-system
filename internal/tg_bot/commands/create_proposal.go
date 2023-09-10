package commands

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/services"
	tgbot "access_governance_system/internal/tg_bot/extension"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const (
	createProposalCommandName = "create_proposal"

	waitingForTypeState     = "waiting_for_type"
	waitingForNicknameState = "waiting_for_nickname"
	waitingForNameState     = "waiting_for_name"
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
	voteService        services.VoteService
	logger             *zap.SugaredLogger
}

func NewCreateProposalCommand(
	appConfig configs.App,
	userRepository repositories.UserRepository,
	proposalRepository repositories.ProposalRepository,
	voteService services.VoteService,
	logger *zap.SugaredLogger,
) Command {
	return &createProposalCommand{
		appConfig:          appConfig,
		userRepository:     userRepository,
		proposalRepository: proposalRepository,
		voteService:        voteService,
		logger:             logger,
	}
}

func (c *createProposalCommand) CanHandle(command string) bool {
	return command == createProposalCommandName
}

func (c *createProposalCommand) Handle(command string, user *models.User, chatID int64) []tgbotapi.Chattable {
	if command == createProposalCommandName {
		return []tgbotapi.Chattable{c.handleCreateProposalCommand(user, chatID)}
	}

	switch user.TelegramState.LastCommandState {
	case waitingForTypeState:
		return []tgbotapi.Chattable{c.handleWaitingForTypeState(command, user, chatID)}
	case waitingForNicknameState:
		return []tgbotapi.Chattable{c.handleWaitingForNicknameState(command, user, chatID)}
	case waitingForNameState:
		return []tgbotapi.Chattable{c.handleWaitingForNameState(command, user, chatID)}
	case waitingForReasonState:
		return []tgbotapi.Chattable{c.handleWaitingForReasonState(command, user, chatID)}
	case waitingForConfirmState:
		return []tgbotapi.Chattable{c.handleWaitingForConfirmState(command, user, chatID)}
	default:
		c.logger.Errorf("user has unknown state: %s", user.TelegramState.LastCommandState)
		return []tgbotapi.Chattable{tgbot.DefaultErrorMessage(chatID)}
	}
}

func (c *createProposalCommand) handleCreateProposalCommand(user *models.User, chatID int64) tgbotapi.Chattable {
	user.TempProposal = models.Proposal{NominatorID: user.ID}

	switch user.Role {
	case models.UserRoleMember:
		return c.handleWaitingForTypeState(models.UserRoleMember.CapitalizedString(), user, chatID)
	case models.UserRoleSeeder:
		parseMode := tgbotapi.ModeMarkdownV2
		messageText := tgbotapi.EscapeText(parseMode, "Кого ты хочешь добавить - member или seeder?")
		messageText = strings.Replace(messageText, "member", "*member*", -1)
		messageText = strings.Replace(messageText, "seeder", "*seeder*", -1)
		message := tgbotapi.NewMessage(chatID, messageText)
		message.ParseMode = parseMode
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
	message := tgbotapi.NewMessage(chatID, "Окей, теперь напиши никнейм пользователя в telegram, которого ты хочешь добавить в сообщество.")

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
	proposalNomineeNickname = strings.TrimPrefix(proposalNomineeNickname, "@")

	proposals, err := c.proposalRepository.GetManyByNomineeNickname(proposalNomineeNickname)
	if err != nil {
		c.logger.Errorf("failed to get proposals by nominee nickname: %s", err)
		return tgbot.DefaultErrorMessage(chatID)
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
			return tgbotapi.NewMessage(chatID, "Предыдущее предложение на добавление этого участника в сообщество ещё не рассмотрено.")
		case models.ProposalStatusApproved:
			c.logger.Warnf(
				"user tried to create proposal for nominee with existing approved proposal: %s, %d, %s",
				proposalNomineeNickname,
				lastProposal.ID,
				lastProposal.CreatedAt,
			)
			return tgbotapi.NewMessage(chatID, "Этот участник уже состоит в сообществе.")
		case models.ProposalStatusRejected:
			monthsAgo := time.Now().AddDate(0, 0, -c.appConfig.RenominationPeriodDays)

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

	foundUser, err := c.userRepository.GetOneByTelegramNickname(proposalNomineeNickname)
	if err != nil {
		c.logger.Errorf("failed to get user by nominee nickname: %s", err)
		return tgbot.DefaultErrorMessage(chatID)
	} else if foundUser != nil {
		c.logger.Warnf(
			"user tried to create proposal for nominee with existing approved proposal: %s",
			proposalNomineeNickname,
		)
		return tgbotapi.NewMessage(chatID, "Этот участник уже состоит в сообществе.")
	}

	text := fmt.Sprintf(
		"Перед тем, как мы перейдем к следующему шагу, тебе нужно убедиться правильно ли ты ввел никнейм пользователя. "+
			"Для это просто нажми на него @%s (если на него нельзя нажать, то похоже ты ввел несуществующий никнейм).\n\n"+
			"Если вдруг ты ошибся, то ты всегда можешь начать сначала вызвав команду /cancel_proposal.\n\n"+
			"Напиши ФИО человека, которого ты хочешь добавить.", proposalNomineeNickname,
	)
	message := tgbotapi.NewMessage(chatID, text)

	user.TempProposal.NomineeTelegramNickname = proposalNomineeNickname

	user.TelegramState.LastCommandState = waitingForNameState
	_ = c.updateUser(user)

	return message
}

func (c *createProposalCommand) handleWaitingForNameState(proposalNomineeName string, user *models.User, chatID int64) tgbotapi.Chattable {
	message := tgbotapi.NewMessage(chatID, "Теперь напиши, почему ты считаешь, что этого участника стоит добавить. Чем подробнее, тем лучше мы сможем понять твою точку зрения.")

	user.TempProposal.NomineeName = proposalNomineeName

	user.TelegramState.LastCommandState = waitingForReasonState
	_ = c.updateUser(user)

	return message
}

func (c *createProposalCommand) handleWaitingForReasonState(proposalDescription string, user *models.User, chatID int64) tgbotapi.Chattable {
	user.TempProposal.Comment = proposalDescription

	messageText := ""
	messageText += fmt.Sprintf("Тип: %s\n", user.TempProposal.NomineeRole)
	messageText += fmt.Sprintf("Участник: %s (@%s)\n", user.TempProposal.NomineeName, user.TempProposal.NomineeTelegramNickname)
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
		user.TempProposal = models.Proposal{}
		user.TelegramState = models.TelegramState{LastCommand: createProposalCommandName}
		_ = c.updateUser(user)
		return c.handleCreateProposalCommand(user, chatID)
	}

	createdAt := time.Now()
	finishedAt := createdAt.AddDate(0, 0, c.appConfig.VotingDurationDays)

	title := user.TempProposal.NomineeName
	description := fmt.Sprintf("@%s предлагает добавить @%s в сообщество\n\nКомментарий: %s", user.TelegramNickname, user.TempProposal.NomineeTelegramNickname, user.TempProposal.Comment)
	dueDate := time.Date(finishedAt.Year(), finishedAt.Month(), finishedAt.Day(), 12, 0, 0, 0, finishedAt.Location())
	poll, err := c.voteService.CreatePoll(title, description, dueDate)
	if err != nil {
		c.logger.Errorw("failed to create poll", "error", err)
		return tgbot.DefaultErrorMessage(chatID)
	}

	user.TempProposal.CreatedAt = createdAt
	user.TempProposal.FinishedAt = finishedAt

	if poll != (models.Poll{}) {
		user.TempProposal.Poll = poll
	}

	proposal, err := c.proposalRepository.Create(&user.TempProposal)
	if err != nil {
		c.logger.Errorw("failed to create proposal", "error", err)
		return tgbot.DefaultErrorMessage(chatID)
	}

	c.logger.Info("proposal created")

	user.Proposals = append(user.Proposals, *proposal)

	if err = c.updateUser(user); err != nil {
		c.logger.Errorw("failed to update user", "error", err)
		return tgbot.DefaultErrorMessage(chatID)
	}
	c.logger.Info("user updated")

	message := tgbotapi.NewMessage(chatID, "Предложение отправлено на голосование.")
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
