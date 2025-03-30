package discordbot

import (
	"discord_ladder_bot/internal/config"
	"discord_ladder_bot/internal/rankingdata"
	"discord_ladder_bot/internal/version"
	"fmt"

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
					Name:        "user",
					Type:        discordgo.ApplicationCommandOptionUser,
					Description: "The alternate discord user to register (admin only).",
					Required:    false,
				},
				{
					Name:        "gamename",
					Type:        discordgo.ApplicationCommandOptionString,
					Description: "In game name.",
					Required:    false,
				},
			},
		},
		{
			Name:        "unregister",
			Description: "Unregister a user from a 1v1 ranking tournament.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "user",
					Type:        discordgo.ApplicationCommandOptionUser,
					Description: "The user to unregister (admin only).",
					Required:    false,
				},
			},
		},
		{
			Name:        "challenge",
			Description: "Challenge a user to a for their position.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "user",
					Type:        discordgo.ApplicationCommandOptionUser,
					Description: "The user to challenge.",
					Required:    true,
				},
			},
		},
		{
			Name:        "result",
			Description: "Report a result of a challenge (only valid from the defender/challengee)",
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
			},
		},
		{
			Name:        "cancel",
			Description: "Cancel a challenge.",
		},
		{
			Name:        "forfeit",
			Description: "Forfeit a challenge (alternate to \"/result result:lost\").",
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
					Name:        "user",
					Type:        discordgo.ApplicationCommandOptionUser,
					Description: "Set data for another user (admin only).",
					Required:    false,
				},
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

	return nil
}

// Stop the bot
func (bot *DiscordBot) Stop() {
	for _, command := range bot.commands {
		bot.Discord.ApplicationCommandDelete(bot.Discord.State.User.ID, "", command.ID)
	}
	bot.Discord.Close()
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
				Content: "Unknown slash command: " + command,
			},
		})
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
