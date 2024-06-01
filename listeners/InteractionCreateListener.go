package listeners

import (
	"context"
	"csrvbot/commands"
	"csrvbot/domain/entities"
	"csrvbot/internal/repos"
	"csrvbot/internal/services"
	"csrvbot/pkg"
	"csrvbot/pkg/discord"
	"csrvbot/pkg/logger"
	"database/sql"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type InteractionCreateListener struct {
	GiveawayCommand      commands.GiveawayCommand
	ThxCommand           commands.ThxCommand
	ThxmeCommand         commands.ThxmeCommand
	CsrvbotCommand       commands.CsrvbotCommand
	DocCommand           commands.DocCommand
	ResendCommand        commands.ResendCommand
	GiveawayHours        string
	CraftserveUrl        string
	GiveawayRepo         repos.GiveawayRepo
	MessageGiveawayRepo  repos.MessageGiveawayRepo
	ServerRepo           repos.ServerRepo
	HelperService        services.HelperService
	JoinableGiveawayRepo entities.JoinableGiveawayRepo
}

func NewInteractionCreateListener(giveawayCommand commands.GiveawayCommand, thxCommand commands.ThxCommand, thxmeCommand commands.ThxmeCommand, csrvbotCommand commands.CsrvbotCommand, docCommand commands.DocCommand, resendCommand commands.ResendCommand, giveawayHours, craftserveUrl string, giveawayRepo *repos.GiveawayRepo, messageGiveawayRepo *repos.MessageGiveawayRepo, serverRepo *repos.ServerRepo, helperService *services.HelperService, joinableGiveawayRepo entities.JoinableGiveawayRepo) InteractionCreateListener {
	return InteractionCreateListener{
		GiveawayCommand:      giveawayCommand,
		ThxCommand:           thxCommand,
		ThxmeCommand:         thxmeCommand,
		CsrvbotCommand:       csrvbotCommand,
		DocCommand:           docCommand,
		ResendCommand:        resendCommand,
		GiveawayHours:        giveawayHours,
		CraftserveUrl:        craftserveUrl,
		GiveawayRepo:         *giveawayRepo,
		MessageGiveawayRepo:  *messageGiveawayRepo,
		ServerRepo:           *serverRepo,
		HelperService:        *helperService,
		JoinableGiveawayRepo: joinableGiveawayRepo,
	}
}

func (h InteractionCreateListener) Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := pkg.CreateContext()
	log := logger.GetLoggerFromContext(ctx).WithGuild(i.GuildID)
	if i.Member != nil {
		log = log.WithUser(i.Member.User.ID)
	} else {
		log = log.WithUser(i.User.ID)
	}
	ctx = logger.ContextWithLogger(ctx, log)
	log.Debug("InteractionCreate event received, type: ", i.Type)

	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		h.handleApplicationCommands(ctx, s, i)
	case discordgo.InteractionApplicationCommandAutocomplete:
		h.handleApplicationCommandsAutocomplete(ctx, s, i)
	case discordgo.InteractionMessageComponent:
		h.handleMessageComponents(ctx, s, i)
	}
}

func (h InteractionCreateListener) handleApplicationCommands(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx).WithCommand(i.ApplicationCommandData().Name)
	ctx = logger.ContextWithLogger(ctx, log)
	log.Debug("Command received")
	switch i.ApplicationCommandData().Name {
	case "giveaway":
		h.GiveawayCommand.Handle(ctx, s, i)
	case "thx":
		h.ThxCommand.Handle(ctx, s, i)
	case "thxme":
		h.ThxmeCommand.Handle(ctx, s, i)
	case "doc":
		h.DocCommand.Handle(ctx, s, i)
	case "csrvbot":
		h.CsrvbotCommand.Handle(ctx, s, i)
	case "resend":
		h.ResendCommand.Handle(ctx, s, i)
	}
}

func (h InteractionCreateListener) handleApplicationCommandsAutocomplete(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "doc":
		h.DocCommand.HandleAutocomplete(ctx, s, i)
	}
}

func (h InteractionCreateListener) handleMessageComponents(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)

	switch i.MessageComponentData().CustomID {
	case "thxwinnercode":
		log.Debug("User clicked thxwinnercode button")
		hasWon, err := h.GiveawayRepo.HasWonGiveawayByMessageId(ctx, i.Message.ID, i.Member.User.ID)
		if err != nil {
			log.WithError(err).Errorf("handleMessageComponents#GiveawayRepo.HasWonGiveawayByMessageId: %v", err)
			return
		}
		if !hasWon {
			log.Debug("User has not won the giveaway")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nie wygrałeś tego giveawayu!")
			return
		}

		log.Debug("User has won the giveaway, getting code...")
		code, err := h.GiveawayRepo.GetCodeForInfoMessage(ctx, i.Message.ID)
		if err != nil {
			log.WithError(err).Errorf("handleMessageComponents#GiveawayRepo.GetCodeForInfoMessage: %v", err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{discord.ConstructWinnerEmbed(h.CraftserveUrl, code)},
			},
		})
		if err != nil {
			log.WithError(err).Errorf("handleMessageComponents#session.InteractionRespond: %v", err)
			return
		}
	case "msgwinnercode":
		log.Debug("User clicked msgwinnercode button")
		hasWon, err := h.MessageGiveawayRepo.HasWonGiveawayByMessageId(ctx, i.Message.ID, i.Member.User.ID)
		if err != nil {
			log.WithError(err).Errorf("handleMessageComponents#MessageGiveawayRepo.HasWonGiveawayByMessageId: %v", err)
			return
		}
		if !hasWon {
			log.Debug("User has not won the giveaway")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nie wygrałeś tego giveawayu!")
			return
		}

		log.Debug("User has won the giveaway, getting code...")
		codes, err := h.MessageGiveawayRepo.GetCodesForInfoMessage(ctx, i.Message.ID, i.Member.User.ID)
		if err != nil {
			log.WithError(err).Errorf("handleMessageComponents#MessageGiveawayRepo.GetCodesForInfoMessage: %v", err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{discord.ConstructMessageWinnerEmbed(h.CraftserveUrl, codes)},
			},
		})
		if err != nil {
			log.WithError(err).Errorf("handleMessageComponents#session.InteractionRespond: %v", err)
			return
		}
	case "accept", "reject":
		h.handleAcceptDeclineButtons(ctx, s, i)
	case "giveawayjoin":
		log.Debug("User clicked giveawayjoin button")

		giveaway, err := h.JoinableGiveawayRepo.GetGiveawayByMessageId(ctx, i.Message.ID)
		if err != nil {
			log.WithError(err).Errorf("handleMessageComponents#JoinableGiveawayRepo.GetGiveawayByMessageId: %v", err)
			return
		}

		if giveaway.EndTime != nil {
			log.Debug("SqlGiveaway has ended")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Ten giveaway już się zakończył!")
			return
		}

		participants, err := h.JoinableGiveawayRepo.GetParticipantsForGiveaway(ctx, giveaway.Id)
		if err != nil {
			log.WithError(err).Errorf("handleMessageComponents#JoinableGiveawayRepo.GetParticipantsForGiveaway: %v", err)
			return
		}

		for _, participant := range participants {
			if participant.UserId == i.Member.User.ID {
				log.Debug("User is already a participant")
				discord.RespondWithEphemeralMessage(ctx, s, i, "Już jesteś uczestnikiem tego giveawayu!")
				return
			}
		}

		memberLevel, err := discord.GetMemberLevel(ctx, s, i.Member, i.GuildID)
		if err != nil {
			log.WithError(err).Error("Could not get member level")
			return
		}

		if giveaway.Level != nil {
			if memberLevel < *giveaway.Level {
				log.Debug("User does not have required level")
				discord.RespondWithEphemeralMessage(ctx, s, i, "Nie masz wymaganego poziomu, żeby wziąć udział w tym giveawayu!")
				return
			}
		}

		err = h.JoinableGiveawayRepo.InsertParticipant(ctx, giveaway.Id, memberLevel, i.Member.User.ID, i.Member.User.Username)
		if err != nil {
			log.WithError(err).Errorf("handleMessageComponents#JoinableGiveawayRepo.InsertParticipant: %v", err)
			return
		}

		log.Infof("%s joined joinable giveaway", i.Member.User.Username)
		discord.RespondWithEphemeralMessage(ctx, s, i, "Zostałeś dodany do giveawayu!")

		// Edit embed to count new participant
		log.Debug("Editing message to count new participant...")

		participantsCount, err := h.JoinableGiveawayRepo.CountParticipantsForGiveaway(ctx, giveaway.Id)
		if err != nil {
			log.WithError(err).Errorf("handleMessageComponents#JoinableGiveawayRepo.CountParticipantsForGiveaway: %v", err)
			return
		}

		var embed *discordgo.MessageEmbed
		if giveaway.Level != nil {
			levelRole, err := discord.GetRoleForLevel(ctx, s, i.GuildID, *giveaway.Level)
			if err != nil {
				log.WithError(err).Error("Could not get role for level")
				return
			}
			embed = discord.ConstructJoinableGiveawayEmbed(h.CraftserveUrl, participantsCount, &levelRole.ID)
		} else {
			embed = discord.ConstructJoinableGiveawayEmbed(h.CraftserveUrl, participantsCount, nil)
		}

		_, err = s.ChannelMessageEditEmbed(i.ChannelID, i.Message.ID, embed)
		if err != nil {
			log.WithError(err).Errorf("handleMessageComponents#session.ChannelMessageEditEmbed: %v", err)
			return
		}
	case "joinablewinnercode":
		log.Debug("User clicked joinablewinnercode button")

		hasWon, err := h.JoinableGiveawayRepo.HasWonGiveawayByMessageId(ctx, i.Message.ID, i.Member.User.ID)
		if err != nil {
			log.WithError(err).Errorf("handleMessageComponents#JoinableGiveawayRepo.HasWonGiveawayByMessageId: %v", err)
			return
		}

		if !hasWon {
			log.Debug("User has not won the giveaway")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nie wygrałeś tego giveawayu!")
			return
		}

		log.Debug("User has won the giveaway, getting code...")
		code, err := h.JoinableGiveawayRepo.GetCodeForInfoMessage(ctx, i.Message.ID, i.Member.User.ID)
		if err != nil {
			log.WithError(err).Errorf("handleMessageComponents#JoinableGiveawayRepo.GetCodesForInfoMessage: %v", err)
			return
		}

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:  discordgo.MessageFlagsEphemeral,
				Embeds: []*discordgo.MessageEmbed{discord.ConstructWinnerEmbed(h.CraftserveUrl, code)},
			},
		})
		if err != nil {
			log.WithError(err).Errorf("handleMessageComponents#session.InteractionRespond: %v", err)
			return
		}
	}
}

func (h InteractionCreateListener) handleAcceptDeclineButtons(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	log := logger.GetLoggerFromContext(ctx)
	if i.Message == nil {
		log.Error("Message is nil")
		return
	}
	log = log.WithMessage(i.Message.ID)
	isThxMessage, err := h.GiveawayRepo.IsThxMessage(ctx, i.Message.ID)
	if err != nil {
		log.WithError(err).Errorf("handleAcceptDeclineButtons#h.GiveawayRepo.IsThxMessage: %v", err)
		return
	}
	isThxmeMessage, err := h.GiveawayRepo.IsThxmeMessage(ctx, i.Message.ID)
	if err != nil {
		log.WithError(err).Errorf("handleAcceptDeclineButtons#h.GiveawayRepo.IsThxmeMessage: %v", err)
		return
	}

	if !isThxMessage && !isThxmeMessage {
		log.Debug("Message is not a thx or thxme message")
		return
	}

	componentId := i.MessageComponentData().CustomID

	member := i.Member

	serverConfig, err := h.ServerRepo.GetServerConfigForGuild(ctx, i.GuildID)
	if err != nil {
		log.WithError(err).Error("Could not get server config")
		return
	}

	if isThxMessage {
		log.Debug("Message is a thx message")
		isAdmin := discord.HasAdminPermissions(ctx, s, member, serverConfig.AdminRoleId, i.GuildID)
		if !isAdmin {
			log.Debug("User is not an admin")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nie masz uprawnień do akceptacji!")
			return
		}

		participant, err := h.GiveawayRepo.GetParticipant(ctx, i.Message.ID)
		if err != nil {
			log.WithError(err).Errorf("handleAcceptDeclineButtons#h.GiveawayRepo.GetParticipant: %v", err)
			return
		}

		giveaway, err := h.GiveawayRepo.GetGiveawayById(ctx, participant.GiveawayId)
		if err != nil {
			log.WithError(err).Errorf("handleAcceptDeclineButtons#h.GiveawayRepo.GetGiveawayById: %v", err)
			return
		}

		if giveaway.EndTime != nil {
			log.Debug("SqlGiveaway has ended")
			discord.RespondWithEphemeralMessage(ctx, s, i, "SqlGiveaway już się zakończył!")
			return
		}

		thxNotification, notificationErr := h.GiveawayRepo.GetThxNotification(ctx, i.Message.ID)
		if notificationErr != nil && !errors.Is(notificationErr, sql.ErrNoRows) {
			log.WithError(notificationErr).Errorf("handleAcceptDeclineButtons#h.GiveawayRepo.GetThxNotification: %v", notificationErr)
			return
		}

		switch componentId {
		case "accept":
			log.Debug("User clicked accept button, updating participant...")
			err = h.GiveawayRepo.UpdateParticipant(ctx, &participant, member.User.ID, member.User.Username, true)
			if err != nil {
				log.WithError(err).Errorf("handleAcceptDeclineButtons#h.GiveawayRepo.UpdateParticipant: %v", err)
				return
			}
			discord.RespondWithEphemeralMessage(ctx, s, i, "Udział użytkownika został potwierdzony!")
			log.Infof("%s accepted %s participation in giveaway %d", member.User.Username, participant.UserName, participant.GiveawayId)

			participants, err := h.GiveawayRepo.GetParticipantNamesForGiveaway(ctx, participant.GiveawayId)
			if err != nil {
				log.WithError(err).Errorf("handleAcceptDeclineButtons#h.GiveawayRepo.GetParticipantNamesForGiveaway: %v", err)
				return
			}

			embed := discord.ConstructThxEmbed(h.CraftserveUrl, participants, h.GiveawayHours, participant.UserId, member.User.ID, "confirm")

			_, err = s.ChannelMessageEditEmbed(i.ChannelID, i.Message.ID, embed)
			if err != nil {
				log.WithError(err).Errorf("handleAcceptDeclineButtons#session.ChannelMessageEditEmbed: %v", err)
				return
			}

			if errors.Is(notificationErr, sql.ErrNoRows) {
				notificationMessageId, err := discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, "", i.GuildID, i.ChannelID, i.Message.ID, participant.UserId, member.User.ID, "confirm", h.CraftserveUrl)
				if err != nil {
					log.WithError(err).Error("Could not notify thx on thx info channel")
					return
				}

				log.Debug("Inserting thx notification...")
				err = h.GiveawayRepo.InsertThxNotification(ctx, i.Message.ID, notificationMessageId)
				if err != nil {
					log.WithError(err).Error("Could not insert thx notification")
					return
				}
			} else {
				_, err = discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, thxNotification.NotificationMessageId, i.GuildID, i.ChannelID, i.Message.ID, participant.UserId, member.User.ID, "confirm", h.CraftserveUrl)
				if err != nil {
					log.WithError(err).Error("Could not notify thx on thx info channel")
					return
				}
			}

			log.Debug("Checking if helper role should be given to participant...")
			h.HelperService.CheckHelper(ctx, s, i.GuildID, participant.UserId)
		case "reject":
			log.Debug("User clicked reject button, updating participant...")
			err := h.GiveawayRepo.UpdateParticipant(ctx, &participant, member.User.ID, member.User.Username, false)
			if err != nil {
				log.WithError(err).Errorf("handleAcceptDeclineButtons#h.GiveawayRepo.UpdateParticipant: %v", err)
				return
			}
			discord.RespondWithEphemeralMessage(ctx, s, i, "Udział użytkownika został odrzucony!")
			log.Infof("%s rejected %s participation in giveaway %d", member.User.Username, participant.UserName, participant.GiveawayId)

			participants, err := h.GiveawayRepo.GetParticipantNamesForGiveaway(ctx, participant.GiveawayId)
			if err != nil {
				log.WithError(err).Errorf("handleAcceptDeclineButtons#h.GiveawayRepo.GetParticipantNamesForGiveaway: %v", err)
				return
			}

			embed := discord.ConstructThxEmbed(h.CraftserveUrl, participants, h.GiveawayHours, participant.UserId, member.User.ID, "reject")

			_, err = s.ChannelMessageEditEmbed(i.ChannelID, i.Message.ID, embed)
			if err != nil {
				log.WithError(err).Errorf("handleAcceptDeclineButtons#session.ChannelMessageEditEmbed: %v", err)
				return
			}

			if errors.Is(notificationErr, sql.ErrNoRows) {
				notificationMessageId, err := discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, "", i.GuildID, i.ChannelID, i.Message.ID, participant.UserId, member.User.ID, "reject", h.CraftserveUrl)
				if err != nil {
					log.WithError(err).Error("Could not notify thx on thx info channel")
					return
				}

				log.Debug("Inserting thx notification...")
				err = h.GiveawayRepo.InsertThxNotification(ctx, i.Message.ID, notificationMessageId)
				if err != nil {
					log.WithError(err).Error("Could not insert thx notification")
					return
				}
			} else {
				_, err = discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, thxNotification.NotificationMessageId, i.GuildID, i.ChannelID, i.Message.ID, participant.UserId, member.User.ID, "reject", h.CraftserveUrl)
				if err != nil {
					log.WithError(err).Error("Could not notify thx on thx info channel")
					return
				}
			}

			log.Debug("Checking if helper role should be given to participant...")
			h.HelperService.CheckHelper(ctx, s, i.GuildID, participant.UserId)
		}

	} else if isThxmeMessage {
		log.Debug("Message is a thxme message")
		candidate, err := h.GiveawayRepo.GetParticipantCandidate(ctx, i.Message.ID)
		if err != nil {
			log.WithError(err).Errorf("handleAcceptDeclineButtons#h.GiveawayRepo.GetParticipantCandidate: %v", err)
			return
		}

		if member.User.ID != candidate.CandidateApproverId {
			log.Debug("User is not the approver of the candidate")
			discord.RespondWithEphemeralMessage(ctx, s, i, "Nie masz uprawnień do zmiany statusu tej prośby!")
			return
		}

		switch componentId {
		case "accept":
			log.Debug("User clicked accept button, updating participant candidate...")
			err := h.GiveawayRepo.UpdateParticipantCandidate(ctx, &candidate, true)
			if err != nil {
				log.WithError(err).Errorf("handleAcceptDeclineButtons#h.GiveawayRepo.UpdateParticipantCandidate: %v", err)
				return
			}
			discord.RespondWithEphemeralMessage(ctx, s, i, "Prośba o podziękowanie zaakceptowana!")
			log.Infof("(%s) %s accepted %s request for thx", i.GuildID, member.User.Username, candidate.CandidateName)

			giveaway, err := h.GiveawayRepo.GetGiveawayForGuild(ctx, i.GuildID)
			if err != nil {
				log.WithError(err).Errorf("handleAcceptDeclineButtons#h.GiveawayRepo.GetGiveawayForGuild: %v", err)
				return
			}
			participants, err := h.GiveawayRepo.GetParticipantNamesForGiveaway(ctx, giveaway.Id)
			if err != nil {
				log.WithError(err).Errorf("handleAcceptDeclineButtons#h.GiveawayRepo.GetParticipantNamesForGiveaway: %v", err)
				return
			}

			embed := discord.ConstructThxEmbed(h.CraftserveUrl, participants, h.GiveawayHours, candidate.CandidateId, "", "wait")

			content := "Prośba o podziękowanie zaakceptowana przez: " + member.User.Mention()
			_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
				Channel: i.ChannelID,
				ID:      i.Message.ID,
				Content: &content,
				Embed:   embed,
			})
			if err != nil {
				log.WithError(err).Errorf("handleAcceptDeclineButtons#session.ChannelMessageEditComplex: %v", err)
				return
			}

			guild, err := s.Guild(i.GuildID)
			if err != nil {
				log.WithError(err).Errorf("handleAcceptDeclineButtons#session.Guild: %v", err)
				return
			}

			log.Debug("Inserting participant...")
			err = h.GiveawayRepo.InsertParticipant(ctx, giveaway.Id, guild.ID, guild.Name, candidate.CandidateId, candidate.CandidateName, i.ChannelID, i.Message.ID)
			if err != nil {
				log.WithError(err).Errorf("handleAcceptDeclineButtons#h.GiveawayRepo.InsertParticipant: %v", err)
				str := "Coś poszło nie tak przy dodawaniu podziękowania :("
				_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &str,
				})
				if err != nil {
					log.WithError(err).Errorf("handleAcceptDeclineButtons#session.InteractionResponseEdit: %v", err)
				}
				return
			}

			log.Infof("%s thanked %s", member.User.Username, candidate.CandidateName)

			thxNotification, err := h.GiveawayRepo.GetThxNotification(ctx, i.Message.ID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				log.WithError(err).Error("Could not get thx notification for message %s", i.Message.ID)
				return
			}

			if errors.Is(err, sql.ErrNoRows) {
				notificationMessageId, err := discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, "", i.GuildID, i.ChannelID, i.Message.ID, candidate.CandidateId, "", "wait", h.CraftserveUrl)
				if err != nil {
					log.WithError(err).Error("Could not notify thx on thx info channel")
					return
				}

				log.Debug("Inserting thx notification...")
				err = h.GiveawayRepo.InsertThxNotification(ctx, i.Message.ID, notificationMessageId)
				if err != nil {
					log.WithError(err).Error("Could not insert thx notification")
					return
				}
			} else {
				_, err = discord.NotifyThxOnThxInfoChannel(s, serverConfig.ThxInfoChannel, thxNotification.NotificationMessageId, i.GuildID, i.ChannelID, i.Message.ID, candidate.CandidateId, "", "wait", h.CraftserveUrl)
				if err != nil {
					log.WithError(err).Error("Could not notify thx on thx info channel")
					return
				}
			}
		case "reject":
			log.Debug("User clicked reject button, updating participant candidate...")
			err := h.GiveawayRepo.UpdateParticipantCandidate(ctx, &candidate, false)
			if err != nil {
				log.WithError(err).Errorf("handleAcceptDeclineButtons#h.GiveawayRepo.UpdateParticipantCandidate: %v", err)
				return
			}

			_, err = s.ChannelMessageEdit(i.ChannelID, i.Message.ID, fmt.Sprintf("%s, czy chcesz podziękować użytkownikowi %s? - Odrzucono", member.User.Mention(), candidate.CandidateName))
			if err != nil {
				log.WithError(err).Errorf("handleAcceptDeclineButtons#session.ChannelMessageEdit: %v", err)
				return
			}

			discord.RespondWithEphemeralMessage(ctx, s, i, "Prośba o podziękowanie odrzucona!")
			log.Infof("%s rejected %s request for thx", member.User.Username, candidate.CandidateName)
		}

	}
}
