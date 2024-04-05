package main

import (
	"access_governance_system/configs"
	"access_governance_system/internal/db"
	"access_governance_system/internal/db/models"
	"access_governance_system/internal/db/repositories"
	"access_governance_system/internal/di"
	"access_governance_system/internal/services"
	tgbot "access_governance_system/internal/tg_bot/extension"
	"fmt"
	"github.com/go-co-op/gocron"
	"math"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func main() {
	s := gocron.NewScheduler(time.UTC)

	config, err := configs.LoadProposalStateServiceConfig()
	logger := di.NewLogger(config.Logger.AppName, config.App.Environment, config.Logger.URL)

	if err != nil {
		logger.Fatalw("failed to load config", "error", err)
	}
	logger.Info("config loaded")

	logger.Info("starting db")
	database, err := db.StartDB(config.DB, logger)
	if err != nil {
		logger.Fatalw("failed to start db", "error", err)
	}
	logger.Info("db started")

	s.Cron("10 12 * * *").Do(func() {
		logger.Info("initializing repositories and services")
		userRepository := repositories.NewUserRepository(database)
		proposalRepository := repositories.NewProposalRepository(database)
		voteService := services.NewVoteService(config.VoteAPI.URL)

		logger.Info("getting seeders")
		seeders, err := userRepository.GetManyByRole(models.UserRoleSeeder)
		if err != nil {
			logger.Fatalw("failed to get seeders", "error", err)
		}

		logger.Info("getting proposals")
		proposals, err := proposalRepository.GetManyByStatus(models.ProposalStatusCreated)
		if err != nil {
			logger.Fatalw("failed to get proposals", "error", err)
		}

		proposalsNeedToBeUpdated := getProposalsNeedToBeUpdated(seeders, proposals, voteService, userRepository, config, logger)

		if len(proposalsNeedToBeUpdated) == 0 {
			logger.Info("no proposals to update")
		} else {
			updatedProposals := updateProposals(proposalsNeedToBeUpdated, proposalRepository, voteService, userRepository, logger)

			for _, proposal := range updatedProposals {
				sendNotifications(proposal, userRepository, config, logger)
			}

			logger.Info("proposals updated")
		}
	})

	s.StartBlocking()
}

func getProposalsNeedToBeUpdated(
	seeders []*models.User,
	proposals []*models.Proposal,
	voteService services.VoteService,
	userRepository repositories.UserRepository,
	config configs.ProposalStateServiceConfig,
	logger *zap.SugaredLogger,
) []*models.Proposal {
	var proposalsToUpdate []*models.Proposal

	minRequiredSeedersCount := calculateMinRequiredSeedersCount(len(seeders), config)
	minRequiredYesVotesToOverride := calculateMinRequiredYesVotesToOverride(len(seeders), config)

	for _, proposal := range proposals {
		if proposalIsEligibleForUpdate(proposal, voteService, userRepository, logger, minRequiredSeedersCount, minRequiredYesVotesToOverride, config) {
			proposalsToUpdate = append(proposalsToUpdate, proposal)
		}
	}

	return proposalsToUpdate
}

func proposalIsEligibleForUpdate(
	proposal *models.Proposal,
	voteService services.VoteService,
	userRepository repositories.UserRepository,
	logger *zap.SugaredLogger,
	minRequiredSeedersCount, minRequiredYesVotesToOverride int,
	config configs.ProposalStateServiceConfig,
) bool {
	if proposal.FinishedAt.After(time.Now()) {
		return false
	}

	votes, err := voteService.GetVotes(proposal.Poll.ID)
	if err != nil {
		logger.Errorw("Failed to get votes", "error", err)
		return false
	}

	yesVotes, noVotes := countVotes(votes)
	votedSeedersCount := countVotedSeeders(votes, userRepository, logger)

	proposal.Status = proposalStatus(yesVotes, noVotes, votedSeedersCount, minRequiredSeedersCount, minRequiredYesVotesToOverride, config)

	return true
}

func calculateMinRequiredSeedersCount(totalSeeders int, config configs.ProposalStateServiceConfig) int {
	return int(math.Min(math.Round(float64(totalSeeders)*config.Quorum), config.MaxRequiredSeedersCount))
}

func calculateMinRequiredYesVotesToOverride(totalSeeders int, config configs.ProposalStateServiceConfig) int {
	return int(math.Round(float64(totalSeeders) * config.YesVotesToOvercomeNo))
}

func countVotes(votes []services.Vote) (yesVotes, noVotes int) {
	for _, vote := range votes {
		if vote.Option == "yes" {
			yesVotes++
		} else if vote.Option == "no" {
			noVotes++
		}
	}
	return yesVotes, noVotes
}

func countVotedSeeders(votes []services.Vote, userRepository repositories.UserRepository, logger *zap.SugaredLogger) int {
	votedSeedersCount := 0
	for _, vote := range votes {
		user, err := userRepository.GetOneByTelegramID(vote.UserID)
		if err != nil || user == nil {
			continue
		}
		if user.Role == models.UserRoleSeeder {
			votedSeedersCount++
		}
	}
	return votedSeedersCount
}

func proposalStatus(
	yesVotes, noVotes, votedSeedersCount, minRequiredSeedersCount, minRequiredYesVotesToOverride int,
	config configs.ProposalStateServiceConfig,
) models.ProposalStatus {
	minRequiredYesVotes := int(math.Max(math.Round(float64(yesVotes+noVotes)*config.MinYesVotesPercentage), config.MinRequiredYesVotes))

	if votedSeedersCount < minRequiredSeedersCount {
		return models.ProposalStatusNoQuorum
	} else if yesVotes < minRequiredYesVotes || (noVotes > 0 && yesVotes < minRequiredYesVotesToOverride) {
		return models.ProposalStatusRejected
	} else {
		return models.ProposalStatusApproved
	}
}

func updateProposals(
	proposals []*models.Proposal,
	proposalRepository repositories.ProposalRepository,
	voteService services.VoteService,
	userRepository repositories.UserRepository,
	logger *zap.SugaredLogger,
) []*models.Proposal {
	var updatedProposals []*models.Proposal

	for _, proposal := range proposals {
		_, err := proposalRepository.Update(proposal)
		if err != nil {
			logger.Errorw("failed to update proposal", "error", err)
			continue
		}

		if proposal.Status == models.ProposalStatusApproved {
			votes, err := voteService.GetVotes(proposal.Poll.ID)
			if err != nil {
				logger.Errorw("failed to get votes", "error", err)
				continue
			}

			backersIDs := make([]int64, 0)
			for _, vote := range votes {
				if vote.Option == "yes" {
					backersIDs = append(backersIDs, vote.UserID)
				}
			}

			user := &models.User{
				Name:             proposal.NomineeName,
				TelegramNickname: proposal.NomineeTelegramNickname,
				Role:             models.UserRoleGuest,
				BackersID:        backersIDs,
			}

			_, err = userRepository.Create(user)
			if err != nil {
				logger.Errorw("failed to create user", "error", err)
				continue
			}
		}

		updatedProposals = append(updatedProposals, proposal)
	}

	return updatedProposals
}

func sendNotifications(
	proposal *models.Proposal,
	userRepository repositories.UserRepository,
	config configs.ProposalStateServiceConfig,
	logger *zap.SugaredLogger,
) {
	switch proposal.Status {
	case models.ProposalStatusRejected:
		sendNotificationsIfProposalRejected(proposal, userRepository, config, logger)
	case models.ProposalStatusApproved:
		sendNotificationsIfProposalApproved(proposal, userRepository, config, logger)
	case models.ProposalStatusNoQuorum:
		sendNotificationsIfProposalNoQuorum(proposal, userRepository, config, logger)
	}
}

func sendNotificationsIfProposalRejected(
	proposal *models.Proposal,
	userRepository repositories.UserRepository,
	config configs.ProposalStateServiceConfig,
	logger *zap.SugaredLogger,
) {
	nominator, err := userRepository.GetOneByID(proposal.NominatorID)
	if err != nil {
		logger.Errorw("could not get nominator", "error", err)
	}

	bot, err := tgbotapi.NewBotAPI(config.AccessGovernanceBot.Token)
	if err != nil {
		logger.Errorw("could not create bot", "error", err)
	}

	messages := []tgbotapi.MessageConfig{
		messageForProposalRejectedToNominator(proposal, nominator),
		messageForProposalRejectedToSeedersGroup(proposal),
	}

	for _, message := range messages {
		_, err = bot.Send(message)
		if err != nil {
			logger.Errorw("could not send message", "error", err)
		}
	}
}

func messageForProposalRejectedToNominator(proposal *models.Proposal, nominator *models.User) tgbotapi.MessageConfig {
	text := fmt.Sprintf(
		`
Кандидатура %s (@%s) была отклонена.

_Это значит, что кворум не состоялся. Голосование по заявкам проходит анонимно в закрытой группе из активных участников сообщества, которые являются носителями ДНК. Повторную заявку на добавление этого человека можно отправить через 3 месяца._ 
`,
		proposal.NomineeName,
		proposal.NomineeTelegramNickname,
	)
	message := tgbotapi.NewMessage(nominator.TelegramID, text)
	message.ParseMode = tgbotapi.ModeMarkdown
	return message
}

func messageForProposalRejectedToSeedersGroup(proposal *models.Proposal) tgbotapi.MessageConfig {
	text := fmt.Sprintf("Кандидатура %s (@%s) была отклонена. Повторная заявка может быть создана через 3 месяца.", proposal.NomineeName, proposal.NomineeTelegramNickname)
	message := tgbotapi.NewMessage(int64(proposal.Poll.ChatID), text)
	message.BaseChat.ReplyToMessageID = proposal.Poll.PollMessageID
	return message
}

func sendNotificationsIfProposalApproved(
	proposal *models.Proposal,
	userRepository repositories.UserRepository,
	config configs.ProposalStateServiceConfig,
	logger *zap.SugaredLogger,
) {
	nominator, err := userRepository.GetOneByID(proposal.NominatorID)
	if err != nil {
		logger.Errorw("could not get nominator", "error", err)
		return
	}

	bot, err := tgbotapi.NewBotAPI(config.AccessGovernanceBot.Token)
	if err != nil {
		logger.Errorw("could not create bot", "error", err)
		return
	}

	inviteLink, err := tgbot.CreateChatInviteLink(bot, config.App.MembersChatID, nominator.TelegramNickname, proposal.NomineeTelegramNickname)
	if err != nil {
		logger.Errorw("could not create invite link", "error", err)
		return
	}

	messages := messagesForProposalApprovedToNominator(proposal, nominator, inviteLink)

	for _, message := range messages {
		_, err = bot.Send(message)
		if err != nil {
			logger.Errorw("could not send message", "error", err)
		}
	}
}

func messagesForProposalApprovedToNominator(proposal *models.Proposal, nominator *models.User, inviteLink string) []tgbotapi.MessageConfig {
	return []tgbotapi.MessageConfig{
		func() tgbotapi.MessageConfig {
			text := fmt.Sprintf(`
Кандидатура %s (@%s) была принята.

Перешли ему следующее сообщение:
`, proposal.NomineeName, proposal.NomineeTelegramNickname)
			message := tgbotapi.NewMessage(nominator.TelegramID, text)
			message.DisableWebPagePreview = true
			return message
		}(),
		func() tgbotapi.MessageConfig {
			text := fmt.Sprintf(`
Привет! Хочу тебя пригласить вступить в группу Shmit16. Я являюсь участником этого сообщества, и мне удалось получить одобрение на твое вступление. 

Для того, чтобы войти в группу, перейди по [ссылке](%s) и нажми кнопку "Присоединиться".

_Комьюнити Shmit16 выросло из группы IT-предпринимателей, которые собирались на бизнес-вечера по адресу Шмитовский проезд, 16. Спустя 10 лет сообщество насчитывает сотни людей разных специальностей по всему миру. Участие в сообществе бесплатное. Вступая в чат, тебе открывается доступ к мероприятиям и дискуссиям сообщества — фестивали, ретриты, онлайн и офлайн._ 
`, inviteLink)
			message := tgbotapi.NewMessage(nominator.TelegramID, text)
			message.ParseMode = tgbotapi.ModeMarkdown
			message.DisableWebPagePreview = true
			return message
		}(),
	}
}

func sendNotificationsIfProposalNoQuorum(
	proposal *models.Proposal,
	userRepository repositories.UserRepository,
	config configs.ProposalStateServiceConfig,
	logger *zap.SugaredLogger,
) {
	nominator, err := userRepository.GetOneByID(proposal.NominatorID)
	if err != nil {
		logger.Errorw("could not get nominator", "error", err)
	}

	bot, err := tgbotapi.NewBotAPI(config.AccessGovernanceBot.Token)
	if err != nil {
		logger.Errorw("could not create bot", "error", err)
	}

	messages := []tgbotapi.MessageConfig{
		messageForProposalNoQuorumToNominator(proposal, nominator),
		messageForProposalNoQuorumToSeedersGroup(proposal),
	}

	for _, message := range messages {
		_, err = bot.Send(message)
		if err != nil {
			logger.Errorw("could not send message", "error", err)
		}
	}
}

func messageForProposalNoQuorumToNominator(proposal *models.Proposal, nominator *models.User) tgbotapi.MessageConfig {
	text := fmt.Sprintf("Кандидатура %s (@%s) была отклонена по причине отсутствия кворума.", proposal.NomineeName, proposal.NomineeTelegramNickname)
	message := tgbotapi.NewMessage(int64(nominator.TelegramID), text)
	return message
}

func messageForProposalNoQuorumToSeedersGroup(proposal *models.Proposal) tgbotapi.MessageConfig {
	text := fmt.Sprintf("Кандидатура %s (@%s) была отклонена по причине отсутствия кворума.", proposal.NomineeName, proposal.NomineeTelegramNickname)
	message := tgbotapi.NewMessage(int64(proposal.Poll.ChatID), text)
	message.BaseChat.ReplyToMessageID = proposal.Poll.PollMessageID
	return message
}
