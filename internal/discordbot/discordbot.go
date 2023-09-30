package discordbot

import (
	"discord_ladder_bot/internal/config"
	"discord_ladder_bot/internal/rankingdata"

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
	bot := &DiscordBot{
		Discord:     discord,
		RankingData: rankingDataPtr,
		commands: []*discordgo.ApplicationCommand{
			{
				Name:        "ladder",
				Description: "Interface with the ranking bot.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "help",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Description: "Help is on the way!",
					},
					{
						Name:        "init",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Description: "Initialize a 1v1 ranking tournament.",
					},
					{
						Name:        "delete_tournament",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Description: "Delete a 1v1 ranking tournament.",
					},
					{
						Name:        "register",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Description: "Register a user to a 1v1 ranking tournament.",
						Options: []*discordgo.ApplicationCommandOption{
							{
								Name:        "user",
								Type:        discordgo.ApplicationCommandOptionUser,
								Description: "The user to register.",
								Required:    false,
							},
						},
					},
					{
						Name:        "unregister",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Description: "Unregister a user from a 1v1 ranking tournament.",
						Options: []*discordgo.ApplicationCommandOption{
							{
								Name:        "user",
								Type:        discordgo.ApplicationCommandOptionUser,
								Description: "The user to unregister.",
								Required:    false,
							},
						},
					},
					{
						Name:        "challenge",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Description: "Challenge a user to a 1v1 match.",
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
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Description: "Report a result of a challenge",
						Options: []*discordgo.ApplicationCommandOption{
							{
								Name:        "result",
								Type:        discordgo.ApplicationCommandOptionString,
								Description: "The result of the challenge.",
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
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Description: "Cancel a challenge.",
					},
					{
						Name:        "forfeit",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Description: "Forfeit a challenge.",
					},
					{
						Name:        "move",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Description: "Move a user to a different position in the ladder.",
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
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Description: "Get the current standings.",
					},
					{
						Name:        "history",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Description: "Get the recent history of matches.",
					},
					{
						Name:        "printraw",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Description: "Print the raw data for the channel.",
					},
					{
						Name:        "set",
						Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
						Description: "Set a value in the ranking data.",
						Options: []*discordgo.ApplicationCommandOption{
							{
								Name:        "user",
								Type:        discordgo.ApplicationCommandOptionSubCommand,
								Description: "Set user data.",
								Options: []*discordgo.ApplicationCommandOption{
									{
										Name:        "user",
										Type:        discordgo.ApplicationCommandOptionUser,
										Description: "The user to set data for.",
										Required:    false,
									},
									{
										Name:        "key",
										Type:        discordgo.ApplicationCommandOptionString,
										Description: "The key to set.",
										Required:    false,
										Choices: []*discordgo.ApplicationCommandOptionChoice{
											{
												Name:  "notes",
												Value: "notes",
											},
											{
												Name:  "status",
												Value: "status",
											},
										},
									},
									{
										Name:        "value",
										Type:        discordgo.ApplicationCommandOptionString,
										Description: "The value to set.",
										Required:    false,
									},
								},
							},
							{
								Name:        "system",
								Type:        discordgo.ApplicationCommandOptionSubCommand,
								Description: "Set system data.",
								Options: []*discordgo.ApplicationCommandOption{
									{
										Name:        "key",
										Type:        discordgo.ApplicationCommandOptionString,
										Description: "The key to set.",
										Required:    false,
										Choices: []*discordgo.ApplicationCommandOptionChoice{
											{
												Name:  "mode",
												Value: "mode",
											},
											{
												Name:  "timeout",
												Value: "timeout",
											},
											{
												Name:  "admin_add",
												Value: "admin_add",
											},
											{
												Name:  "admin_remove",
												Value: "admin_remove",
											},
										},
									},
									{
										Name:        "value",
										Type:        discordgo.ApplicationCommandOptionString,
										Description: "The value to set.",
										Required:    false,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	bot.handlers = map[string]commandHandler{

		"help": func(c *rankingdata.ChannelRankingData,
			i *discordgo.InteractionCreate,
			o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
			return "if this works, maybe you don't really need help...", nil
		},
		"init": func(c *rankingdata.ChannelRankingData,
			i *discordgo.InteractionCreate,
			o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
			if c != nil {
				return "Channel already initialized. If you'd like to reset, use !delete_tournament and then !init.", nil
			} else {
				err := bot.RankingData.AddChannel(i.ChannelID, i.Member.User.ID)
				if err != nil {
					return "", err
				}
			}
			return "Channel initialized!", nil
		},
		"delete_tournament": func(c *rankingdata.ChannelRankingData,
			i *discordgo.InteractionCreate,
			o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
			err := bot.RankingData.RemoveChannel(i.ChannelID)
			if err != nil {
				return "", err
			}
			return "Channel deleted!", nil
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
			return c.PrintLadder()
		},
		"active_challenges": func(c *rankingdata.ChannelRankingData,
			i *discordgo.InteractionCreate,
			o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
			return c.PrintChallenges()
		},
		"history": func(c *rankingdata.ChannelRankingData,
			i *discordgo.InteractionCreate,
			o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
			return c.PrintHistory()
		},
		"set": handleSet,
		"printraw": func(c *rankingdata.ChannelRankingData,
			i *discordgo.InteractionCreate,
			o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
			return c.PrintRaw()
		},
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

	// Register commands
	for _, command := range bot.commands {
		bot.Discord.ApplicationCommandCreate(bot.Discord.State.User.ID, "", command)
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
		_, _ = s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" I'm a bot! Try /ladder to do stuff.")
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

	if data.Name != "ladder" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unknown command: " + data.Name,
			},
		})
		return
	}

	if len(data.Options) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "No subcommand specified.",
			},
		})
		return
	}

	// get the subcommand
	subcommand := data.Options[0].Name

	handler, ok := bot.handlers[subcommand]
	if !ok {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unknown subcommand: " + subcommand,
			},
		})
		return
	}

	// get the channel data
	channel, err := bot.RankingData.FindChannel(i.ChannelID)
	if err != nil && subcommand != "init" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})
		return
	}

	// call the handler
	response, err2 := handler(channel, i, data.Options[0].Options)
	if err2 != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err2.Error(),
			},
		})
		return
	}

	// respond
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
		},
	})

	// save early and often?
	bot.RankingData.Write()

}
