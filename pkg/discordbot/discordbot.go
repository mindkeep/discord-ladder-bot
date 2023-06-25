package discordbot

import (
	"discord_ladder_bot/pkg/config"
	"discord_ladder_bot/pkg/rankingdata"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Names       []string
	Description string
	Usage       string
	Handler     func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error)
}

type DiscordBot struct {
	Discord     *discordgo.Session
	RankingData *rankingdata.RankingData
	commandMap  map[string]*Command
	commands    []*Command
}

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
				response += "\t" + command.Description + "\n"
			}
			return response, nil
		})

	bot.registerCommand(
		[]string{"init"},
		"Initialize a 1v1 ranking tournament.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			if c != nil {
				return "Channel already initialized. If you'd like to reset, use !delete_tournament and then !init.", nil
			} else {
				err := bot.RankingData.AddChannel(m.ChannelID)
				if err != nil {
					return "", err
				}
			}
			return "Channel initialized!", nil
		})

	bot.registerCommand(
		[]string{"register", "join", "add"},
		"Register a user to a 1v1 ranking tournament.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			var playerID string
			if len(m.Mentions) > 1 {
				return "You can only register one user at a time.", nil
			} else if len(m.Mentions) == 1 {
				playerID = m.Mentions[0].ID
			} else {
				playerID = m.Message.Author.ID
			}
			err := c.AddPlayer(playerID)
			if err != nil {
				return "", err
			}
			return "Registered!", nil
		})

	bot.registerCommand(
		[]string{"unregister", "leave", "remove", "quit"},
		"Unregister a user from a 1v1 ranking tournament.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			var playerID string
			if len(m.Mentions) > 1 {
				return "You can only unregister one user at a time.", nil
			} else if len(m.Mentions) == 1 {
				playerID = m.Mentions[0].ID
			} else {
				playerID = m.Message.Author.ID
			}
			err := c.RemovePlayer(playerID)
			if err != nil {
				return "", err
			}
			return "Unregistered!", nil
		})

	bot.registerCommand(
		[]string{"challenge"},
		"Challenge a 1v1 ranking tournament.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			if len(m.Mentions) != 1 {
				return "Please @ mention one person to challenge.", err
			} else {
				err := c.StartChallenge(m.Author.ID, m.Mentions[0].ID)
				if err != nil {
					return "", err
				}
			}
			return "Challenge started!", nil
		})

	bot.registerCommand(
		[]string{"result"},
		"Report a result of a challenge using one of: w, win, l, lost, f, forfeit.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			words := strings.Split(m.Content, " ")
			if len(words) != 2 {
				return "Please use one of: w, win, l, lost, f, forfeit.", nil
			}
			result := words[1]
			switch result {
			case "w":
				result = "win"
			case "l":
				result = "lost"
			case "f":
				result = "forfeit"
			}
			err := c.ResolveChallenge(m.Author.ID, result)
			if err != nil {
				return "", err
			}
			return "Challenge has been resolved... somehow, TODO, add something clever...", nil
		})

	bot.registerCommand(
		[]string{"delete_tournament"},
		"Delete a 1v1 ranking tournament.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			err := bot.RankingData.RemoveChannel(m.ChannelID)
			if err != nil {
				return "", err
			}
			return "Channel deleted!", nil
		})

	bot.registerCommand(
		[]string{"cancel"},
		"Cancel a challenge",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			err := c.ResolveChallenge(m.Author.ID, "cancel")
			if err != nil {
				return "", err
			}
			return "Challenge canceled!", nil
		})

	bot.registerCommand(
		[]string{"ladder"},
		"Display the current tournament ladder state.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			return "TODO", nil
		})

	bot.registerCommand(
		[]string{"active", "challenges"},
		"Display the current active challenges.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			return "TODO", nil
		})

	bot.registerCommand(
		[]string{"set"},
		"Print or adjust game settings.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			return "TODO", nil
		})

	return bot, nil
}

// private function to register a command and map aliases
func (bot *DiscordBot) registerCommand(names []string, description string,
	handler func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error)) {
	// create a command struct
	command := Command{
		Names:       names,
		Description: description,
		Handler:     handler,
	}
	// add the command to the list of commands
	bot.commands = append(bot.commands, &command)
	//map all the commands to the command struct
	for _, name := range names {
		bot.commandMap[name] = &command
	}
}

func (bot *DiscordBot) Start() error {
	bot.Discord.AddHandler(bot.handleMessageCreate)
	err := bot.Discord.Open()
	if err != nil {
		return err
	}
	return nil
}

func (bot *DiscordBot) Stop() {
	bot.Discord.Close()
}

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
		command := words[0][1:]

		// find the channel in the ranking data, or initialize it if the command is "init"
		channel, err := bot.RankingData.FindChannel(m.ChannelID)
		if err != nil && command != "init" && command != "help" {
			_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		// check if the command exists
		if _, ok := bot.commandMap[command]; !ok {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Unknown command.")
			command = "help"
			if _, help_found := bot.commandMap[command]; !help_found {
				_, _ = s.ChannelMessageSend(m.ChannelID, "And there is no help for you either! (sounds like a bad day...)")
				return
			}
		}
		// execute the command
		response, err := bot.commandMap[command].Handler(channel, m)
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID,
				command+" returned with the following error:\n"+err.Error())
			return
		} else {
			_, _ = s.ChannelMessageSend(m.ChannelID, response)
		}

		// save early and often?
		bot.RankingData.Write()
	}
}
