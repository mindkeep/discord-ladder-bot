package discordbot

import (
	"discord_ladder_bot/internal/config"
	"discord_ladder_bot/internal/rankingdata"
	"discord_ladder_bot/internal/version"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

type commandHandler func(*rankingdata.ChannelRankingData,
	*discordgo.InteractionCreate,
	[]*discordgo.ApplicationCommandInteractionDataOption) (string, error)

type DiscordBot struct {
	Discord     *discordgo.Session
	RankingData *rankingdata.RankingData
	commands    []*discordgo.ApplicationCommand
	handlers    map[string]commandHandler
}

// NewDiscordBot creates a new DiscordBot instance
func NewDiscordBot(conf *config.Config) (*DiscordBot, error) {
	discord, err := discordgo.New("Bot " + conf.DiscordToken)

	if err != nil {
		return nil, err
	}
	//discord.LogLevel = discordgo.LogInformational

	rankingDataPtr, err := rankingdata.ReadRankingData(conf)
	if err != nil {
		return nil, err
	}

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "help",
			Description: "Help is on the way!",
		},
		{
			Name:        "init",
			Description: "Initialize a 1v1 ranking tournament (one per channel).",
		},
		{
			Name:        "delete_tournament",
			Description: "Delete a 1v1 ranking tournament (admin only).",
		},
		{
			Name:        "register",
			Description: "Register a user to a 1v1 ranking tournament.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "gamename",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "In game username.",
					Required:    true,
				},
				{
					Name:        "alt_user",
					Type:        discordgo.ApplicationCommandOptionUser,
					Description: "The alternate discord user to register (admin only).",
					Required:    false,
				},
			},
		},
		{
			Name:        "unregister",
			Description: "Unregister a user from a 1v1 ranking tournament.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "alt_user",
					Type:        discordgo.ApplicationCommandOptionUser,
					Description: "The alternate discord user to unregister (admin only).",
					Required:    false,
				},
			},
		},
		{
			Name:        "challenge",
			Description: "Challenge a user to a for their position.",
			Options: []*discordgo.ApplicationCommandOption{

				{
					Name:        "defender",
					Type:        discordgo.ApplicationCommandOptionUser,
					Description: "The user being challenged.",
					Required:    true,
				},
				{
					Name:        "alt_user",
					Type:        discordgo.ApplicationCommandOptionUser,
					Description: "The alternate challenger (admin only).",
					Required:    false,
				},
			},
		},
		{
			Name:        "result",
			Description: "Report a result of a challenge (only valid from the defender)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "result",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "The result of the challenge (defender won or lost).",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "won",
							Value: "won",
						},
						{
							Name:  "lost",
							Value: "lost",
						},
					},
				},
				{
					Name:        "alt_user",
					Type:        discordgo.ApplicationCommandOptionUser,
					Description: "The alternate defender (admin only).",
					Required:    false,
				},
			},
		},
		{
			Name:        "cancel",
			Description: "Cancel a challenge.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "alt_user",
					Type:        discordgo.ApplicationCommandOptionUser,
					Description: "The alternate challenger (admin only).",
					Required:    false,
				},
			},
		},
		{
			Name:        "forfeit",
			Description: "Forfeit a challenge (alternate to \"/result result:lost\").",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "alt_user",
					Type:        discordgo.ApplicationCommandOptionUser,
					Description: "The alternate challenger (admin only).",
					Required:    false,
				},
			},
		},
		{
			Name:        "move",
			Description: "Move a user to a different position in the ladder. (admin only)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "user",
					Type:        discordgo.ApplicationCommandOptionUser,
					Description: "The user to move.",
					Required:    true,
				},
				{
					Name:        "position",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Description: "The position to move the user to.",
					Required:    true,
				},
			},
		},
		{
			Name:        "standings",
			Description: "Get the current standings.",
		},
		{
			Name:        "active_challenges",
			Description: "Get the current active challenges.",
		},
		{
			Name:        "history",
			Description: "Get the recent history of matches.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "limit",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Description: "The number of matches to show (default: 10, not implemented).",
					Required:    false,
				},
			},
		},
		{
			Name:        "user_settings",
			Description: "Set a value in the ranking data.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "gamename",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "In game name.",
					Required:    false,
				},
				{
					Name:        "status",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "The status to set.",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "active",
							Value: "active",
						},
						{
							Name:  "inactive",
							Value: "inactive",
						},
					},
				},
				{
					Name:        "notes",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "Notes to set.",
					Required:    false,
				},
				{
					Name:        "alt_user",
					Type:        discordgo.ApplicationCommandOptionUser,
					Description: "Set data for another user (admin only).",
					Required:    false,
				},
			},
		},
		{
			Name:        "system_settings",
			Description: "Set system data (admin only).",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "mode",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "The challenge mode to set.",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "ladder",
							Value: "ladder",
						},
						{
							Name:  "pyramid",
							Value: "pyramid",
						},
						{
							Name:  "open",
							Value: "open",
						},
					},
				},
				{
					Name:        "timeout",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Description: "The challenge timeout in days to set.",
					Required:    false,
				},
				{
					Name:        "admin_add",
					Type:        discordgo.ApplicationCommandOptionUser,
					Description: "Add an admin.",
					Required:    false,
				},
				{
					Name:        "admin_remove",
					Type:        discordgo.ApplicationCommandOptionUser,
					Description: "Remove an admin.",
					Required:    false,
				},
				{
					Name:        "notes",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "Notes to set.",
					Required:    false,
				},
			},
		},
		{
			Name:        "printraw",
			Description: "Print the raw data for the channel.",
		},
	}

	handlers := map[string]commandHandler{

		"help": func(c *rankingdata.ChannelRankingData,
			i *discordgo.InteractionCreate,
			o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
			var response string
			response += "Commands:\n"
			for _, cmd := range commands {
				response += fmt.Sprintf("  /%s: %s\n", cmd.Name, cmd.Description)
			}
			response += fmt.Sprintf("Version: %s\n", version.Version)
			return response, nil
		},
		"init": func(c *rankingdata.ChannelRankingData,
			i *discordgo.InteractionCreate,
			o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
			if c != nil {
				return "Channel already initialized. If you'd like to reset, use !delete_tournament and then !init.", nil
			}
			return rankingDataPtr.AddChannel(i.ChannelID, i.Member.User.ID)
		},
		"delete_tournament": func(c *rankingdata.ChannelRankingData,
			i *discordgo.InteractionCreate,
			o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
			if !c.IsAdmin(i.Member.User.ID) {
				return "You must be an admin to delete the tournament!", nil
			}
			return rankingDataPtr.RemoveChannel(i.ChannelID)
		},
		"register":   handleRegister,
		"unregister": handleUnregister,
		"challenge":  handleChallenge,
		"result":     handleResult,
		"cancel":     handleCancel,
		"forfeit":    handleForfeit,
		"move":       handleMove,
		"standings": func(c *rankingdata.ChannelRankingData,
			i *discordgo.InteractionCreate,
			o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
			return c.PrintRankings()
		},
		"active_challenges": func(c *rankingdata.ChannelRankingData,
			i *discordgo.InteractionCreate,
			o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
			return c.PrintChallenges()
		},
		"history": func(c *rankingdata.ChannelRankingData,
			i *discordgo.InteractionCreate,
			o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
			// TODO handle limit
			return c.PrintHistory()
		},
		"user_settings":   handleUserSettings,
		"system_settings": handleSystemSettings,
		"printraw": func(c *rankingdata.ChannelRankingData,
			i *discordgo.InteractionCreate,
			o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
			return c.PrintRaw()
		},
	}

	bot := &DiscordBot{
		Discord:     discord,
		RankingData: rankingDataPtr,
		commands:    commands,
		handlers:    handlers,
	}

	return bot, nil
}

// Start the bot
func (bot *DiscordBot) Start() error {

	// Add handlers
	bot.Discord.AddHandler(bot.handleMessageCreate)
	bot.Discord.AddHandler(bot.handleInteractionCreate)

	// Open connection to Discord
	err := bot.Discord.Open()
	if err != nil {
		return err
	}

	oldguilds := bot.Discord.State.Guilds
	for _, guild := range oldguilds {
		fmt.Println("Found guild: ", guild.Name)
		gcmds, err := bot.Discord.ApplicationCommands(bot.Discord.State.User.ID, guild.ID)
		if err != nil {
			for _, cmd := range gcmds {
				fmt.Println("Deleting old command: ", cmd.Name, " in ", guild.Name)
				bot.Discord.ApplicationCommandDelete(bot.Discord.State.User.ID, guild.ID, cmd.ID)
			}
		}
	}

	oldcmds, err := bot.Discord.ApplicationCommands(bot.Discord.State.User.ID, "")
	if err != nil {
		for _, cmd := range oldcmds {
			fmt.Println("Deleting old command: ", cmd.Name)
			bot.Discord.ApplicationCommandDelete(bot.Discord.State.User.ID, "", cmd.ID)
		}
	}

	for _, cmd := range bot.commands {
		fmt.Println("Creating command: ", cmd.Name)
		_, err := bot.Discord.ApplicationCommandCreate(bot.Discord.State.User.ID, "", cmd)
		if err != nil {
			return err
		}
	}

	// Start the nightly cleanup task
	bot.StartNightlyCleanupTask()

	return nil
}

// Stop the bot
func (bot *DiscordBot) Stop() {
	for _, command := range bot.commands {
		bot.Discord.ApplicationCommandDelete(bot.Discord.State.User.ID, "", command.ID)
	}
	bot.Discord.Close()
}

// StartNightlyCleanupTask starts a goroutine that runs cleanup tasks periodically
func (bot *DiscordBot) StartNightlyCleanupTask() {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		// Run cleanup immediately on start
		bot.runCleanupTasks()

		// Then run every 24 hours
		for range ticker.C {
			bot.runCleanupTasks()
		}
	}()
}

// runCleanupTasks performs all cleanup operations
func (bot *DiscordBot) runCleanupTasks() {
	fmt.Println("Running nightly cleanup tasks...")

	totalExpired := 0
	totalRemoved := 0

	// Get all guilds the bot is in
	guilds := bot.Discord.State.Guilds

	for _, channel := range bot.RankingData.Channels {
		// Clean up expired challenges
		expiredCount, err := channel.CleanupExpiredChallenges()
		if err != nil {
			fmt.Printf("Error cleaning up expired challenges for channel %s: %v\n", channel.ChannelID, err)
			continue
		}
		totalExpired += expiredCount

		// Find the guild for this channel
		var guildID string
		for _, guild := range guilds {
			// Try to get the channel from the guild to verify it belongs to this guild
			discordChannel, err := bot.Discord.Channel(channel.ChannelID)
			if err == nil && discordChannel.GuildID == guild.ID {
				guildID = guild.ID
				break
			}
		}

		if guildID == "" {
			// Could not find guild for this channel, skip member check
			continue
		}

		// Get all guild members (with pagination for large guilds)
		memberIDs := make(map[string]bool)
		after := ""
		for {
			members, err := bot.Discord.GuildMembers(guildID, after, 1000)
			if err != nil {
				fmt.Printf("Error getting guild members for guild %s: %v\n", guildID, err)
				break
			}

			if len(members) == 0 {
				break
			}

			// Add members to the map
			for _, member := range members {
				memberIDs[member.User.ID] = true
				after = member.User.ID // Update for next page
			}

			// If we got fewer than 1000, we've reached the end
			if len(members) < 1000 {
				break
			}
		}

		// Check each player and remove if they left the server
		playerIDs := channel.GetAllPlayerIDs()
		for _, playerID := range playerIDs {
			removed, err := channel.RemovePlayerIfNotInGuild(playerID, memberIDs)
			if err != nil {
				fmt.Printf("Error removing player %s: %v\n", playerID, err)
				continue
			}
			if removed {
				totalRemoved++
				fmt.Printf("Removed player %s from channel %s (left server)\n", playerID, channel.ChannelID)
			}
		}
	}

	// Save changes if any cleanup was done
	if totalExpired > 0 || totalRemoved > 0 {
		fmt.Printf("Cleanup complete: %d challenges expired, %d players removed\n", totalExpired, totalRemoved)
		bot.RankingData.Write()
	} else {
		fmt.Println("Cleanup complete: no changes needed")
	}
}

// Handle a message create event
func (bot *DiscordBot) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		//fmt.Println("Ignoring message from self")
		return
	}

	// If the message is addressed to the bot, then respond.
	if len(m.Mentions) > 0 && m.Mentions[0].ID == s.State.User.ID {
		_, _ = s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" I'm a bot! Try /help to find other valid slash commands.")
		return
	}
}

// Handle a slash command
func (bot *DiscordBot) handleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {

	if i.Type == discordgo.InteractionPing {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponsePong,
		})
		return
	}

	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	if i.Member == nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "This bot is intended to be used from a Server.",
			},
		})
		return
	}

	// get the command data
	data := i.ApplicationCommandData()

	// get the subcommand
	command := data.Name

	handler, ok := bot.handlers[command]
	if !ok {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Invalid slash command: " + command,
			},
		})
		err := bot.Discord.ApplicationCommandDelete(bot.Discord.State.User.ID, "", data.ID)
		if err != nil {
			fmt.Println("Error deleting invalid command: ", err)
		} else {
			fmt.Println("Deleted invalid command: ", command)
		}
		return
	}

	// get the channel data
	channel, err := bot.RankingData.FindChannel(i.ChannelID)
	if err != nil && command != "init" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})
		return
	}

	// call the handler
	response, err2 := handler(channel, i, data.Options)
	if err2 != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err2.Error(),
			},
		})
		return
	}

	// determine if we should limit mentions in noisy output commands
	if command == "standings" || command == "active_challenges" || command == "history" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{},
				},
			},
		})
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	}

	// save early and often?
	bot.RankingData.Write()
}
