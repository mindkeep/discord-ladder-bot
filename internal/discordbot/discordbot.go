package discordbot

import (
	"discord_ladder_bot/internal/config"
	"discord_ladder_bot/internal/rankingdata"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Names       []string
	Description string
	Usage       string
	Handler     func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error)
	Privledged  bool
}

type DiscordBot struct {
	Discord     *discordgo.Session
	RankingData *rankingdata.RankingData
	commandMap  map[string]*Command
	commands    []*Command
}

// private function to register a command and map aliases
func (bot *DiscordBot) registerCommand(names []string, description string,
	handler func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error),
	privledged bool) {
	// create a command struct
	command := Command{
		Names:       names,
		Description: description,
		Handler:     handler,
		Privledged:  privledged,
	}
	// add the command to the list of commands
	bot.commands = append(bot.commands, &command)
	//map all the commands to the command struct
	for _, name := range names {
		bot.commandMap[name] = &command
	}
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
		commandMap:  make(map[string]*Command),
		commands:    make([]*Command, 0),
	}

	bot.registerCommand(
		[]string{"help"},
		"Help is on the way!",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			var response string
			for _, command := range bot.commands {
				response += "!" + strings.Join(command.Names, " !") + "\n"
				if command.Privledged {
					response += "\t\t(admin only)\n"
				}
				response += "\t" + command.Description + "\n"
			}
			return response, nil
		},
		false)

	bot.registerCommand(
		[]string{"init"},
		"Initialize a 1v1 ranking tournament.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			if c != nil {
				return "Channel already initialized. If you'd like to reset, use !delete_tournament and then !init.", nil
			} else {
				err := bot.RankingData.AddChannel(m.ChannelID, m.Author.ID)
				if err != nil {
					return "", err
				}
			}
			return "Channel initialized!", nil
		},
		false)
	bot.registerCommand(
		[]string{"delete_tournament"},
		"Delete a 1v1 ranking tournament.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			err := bot.RankingData.RemoveChannel(m.ChannelID)
			if err != nil {
				return "", err
			}
			return "Channel deleted!", nil
		},
		true)

	bot.registerCommand(
		[]string{"register", "join", "add"},
		"Register a user to a 1v1 ranking tournament.",
		handle_register,
		false)

	bot.registerCommand(
		[]string{"unregister", "leave", "remove", "quit"},
		"Unregister a user from a 1v1 ranking tournament.",
		handle_unregister,
		false)

	bot.registerCommand(
		[]string{"challenge"},
		"Challenge a 1v1 ranking tournament.",
		handle_challenge,
		false)

	bot.registerCommand(
		[]string{"result", "results"},
		"Report a result of a challenge using one of: w, won, l, lost, f, forfeit.",
		handle_result,
		false)

	bot.registerCommand(
		[]string{"cancel"},
		"Cancel a challenge",
		handle_cancel,
		false)

	bot.registerCommand(
		[]string{"forfeit"},
		"Forfeit a challenge",
		handle_forfeit,
		false)

	bot.registerCommand(
		[]string{"move"},
		"Move a player to a new rank.",
		handle_move,
		true)

	bot.registerCommand(
		[]string{"ladder"},
		"Display the current tournament ladder state.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			return c.PrintLadder()
		},
		false)

	bot.registerCommand(
		[]string{"active", "challenges"},
		"Display the current active challenges.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			return c.PrintChallenges()
		},
		false)

	bot.registerCommand(
		[]string{"history"},
		"Display the history of the tournament.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			return c.PrintHistory()
		},
		false)

	bot.registerCommand(
		[]string{"set"},
		"Print or adjust game settings.",
		handle_set,
		true)

	bot.registerCommand(
		[]string{"printraw"},
		"Print the raw data for the channel.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			return c.PrintRaw()
		},
		false)

	return bot, nil
}

// Start the bot
func (bot *DiscordBot) Start() error {
	bot.Discord.AddHandler(bot.handleMessageCreate)
	err := bot.Discord.Open()
	if err != nil {
		return err
	}
	return nil
}

// Stop the bot
func (bot *DiscordBot) Stop() {
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
		_, _ = s.ChannelMessageSend(m.ChannelID, m.Author.Mention()+" I'm a bot!")
		return
	}

	// If the message begins to a !, then try to map it to an action.
	if strings.HasPrefix(m.Content, "!") {

		// split message content into word list
		words := strings.Split(strings.ToLower(m.Content), " ")
		command_key := words[0][1:]

		authorIsAdmin := false
		// find the channel in the ranking data, or initialize it if the command is "init"
		channel, err := bot.RankingData.FindChannel(m.ChannelID)
		if err != nil {
			if command_key != "init" && command_key != "help" {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
		} else {
			// channel exists, check if the user is an admin
			authorIsAdmin = channel.IsAdmin(m.Author.ID)
		}

		// find the command
		command, ok := bot.commandMap[command_key]
		if !ok {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Unknown command.")
			command_key = "help"
			if _, help_found := bot.commandMap[command_key]; !help_found {
				_, _ = s.ChannelMessageSend(m.ChannelID, "And there is no help for you either! (sounds like a bad day...)")
				return
			}
		}

		// check if the user is allowed to use the command
		if !authorIsAdmin && command.Privledged {
			_, _ = s.ChannelMessageSend(m.ChannelID, "You are not allowed to use this command.")
			return
		}

		// execute the command
		response, err := command.Handler(channel, m)
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID,
				command_key+" returned with the following error:\n"+err.Error())
			return
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, response)
		}

		// save early and often?
		bot.RankingData.Write()
	}
}
