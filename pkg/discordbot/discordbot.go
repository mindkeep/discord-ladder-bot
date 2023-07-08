package discordbot

import (
	"discord_ladder_bot/pkg/config"
	"discord_ladder_bot/pkg/rankingdata"
	"fmt"
	"strconv"
	"strings"
	"time"

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
		},
		false)

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
		},
		false)

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
		},
		false)

	bot.registerCommand(
		[]string{"result", "results"},
		"Report a result of a challenge using one of: w, won, l, lost, f, forfeit.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			words := strings.Split(m.Content, " ")
			if len(words) != 2 {
				return "Please use one of: w, won, l, lost, f, forfeit.", nil
			}
			result := words[1]
			switch result {
			case "w":
				result = "won"
			case "win":
				result = "won"
			case "l":
				result = "lost"
			case "loss":
				result = "lost"
			case "lose":
				result = "lost"
			case "f":
				result = "forfeit"
			}
			err := c.ResolveChallenge(m.Author.ID, result)
			if err != nil {
				return "", err
			}
			return "Challenge has been resolved... somehow, TODO, add something clever...", nil
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
		false)

	bot.registerCommand(
		[]string{"cancel"},
		"Cancel a challenge",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			err := c.ResolveChallenge(m.Author.ID, "cancel")
			if err != nil {
				return "", err
			}
			return "Challenge canceled!", nil
		},
		false)

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
		// TODO: this is a mess, clean it up
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			words := strings.Split(m.Content, " ")
			if len(words) == 1 {
				var response string
				response += "Game settings:\n"
				response += fmt.Sprintf("  ChallengeMode: %s (ladder or pyramid)\n", c.ChallengeMode)
				response += fmt.Sprintf("  ChallengeTimeoutDays: %d\n", c.ChallengeTimeoutDays)
				response += fmt.Sprintf("  Admins: %s\n", strings.Join(c.Admins, ", "))
				return response, nil
			} else if strings.ToLower(words[1]) == "challengemode" {
				if len(words) != 3 || (words[2] != "ladder" && words[2] != "pyramid") {
					return "Please specify a ChallengeMode (ladder or pyramid).", nil
				}
				c.ChallengeMode = words[2]
				return "ChallengeMode set!", nil
			} else if strings.ToLower(words[1]) == "challengetimeoutdays" {
				error_response := "Please specify a ChallengeTimeoutDays (integer)."
				if len(words) != 3 {
					return error_response, nil
				} else {
					timeoutDays, err := strconv.Atoi(words[2])
					if err != nil {
						return error_response, nil
					}
					c.ChallengeTimeoutDays = time.Hour * time.Duration(24*timeoutDays)
					return "ChallengeTimeoutDays set!", nil
				}
			} else if strings.ToLower(words[1]) == "admins" {
				error_response := "Please specify a command (add or remove) and a user."
				if len(words) != 4 || len(m.Mentions) != 1 {
					return error_response, nil
				} else if strings.ToLower(words[2]) == "add" {
					if !c.IsAdmin(m.Author.ID) {
						return "You must be an admin to add admins.", nil
					}
					if len(c.Admins) == 0 || !c.IsAdmin(m.Mentions[0].ID) {
						c.Admins = append(c.Admins, m.Mentions[0].ID)
						return "Admin added!", nil
					} else {
						return "User is already an admin.", nil
					}
				} else if strings.ToLower(words[2]) == "remove" {
					if !c.IsAdmin(m.Author.ID) {
						return "You must be an admin to remove admins.", nil
					}
					if c.IsAdmin(m.Mentions[0].ID) {
						for i, admin := range c.Admins {
							if admin == m.Mentions[0].ID {
								c.Admins = append(c.Admins[:i], c.Admins[i+1:]...)

								return "Admin removed!", nil
							}
						}
						return "User is not an admin.", nil
					} else {
						return "User is not an admin.", nil

					}
				} else {
					return error_response, nil
				}
			} else {
				return "Unknown setting.", nil
			}
		},
		true)

	bot.registerCommand(
		[]string{"move"},
		"Move a player to a new rank.",
		func(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
			error_response := "Please specify a player and a position."
			words := strings.Split(m.Content, " ")
			if len(words) != 3 || len(m.Mentions) != 1 {
				return error_response, nil
			}
			position, err := strconv.Atoi(words[2])
			if err != nil {
				return error_response, nil
			}
			err2 := c.MovePlayer(m.Mentions[0].ID, position)
			if err2 != nil {
				return "", err2
			}
			return "Player moved!", nil
		},
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
