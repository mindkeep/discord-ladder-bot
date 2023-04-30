package discordbot

import (
	"discord_ladder_bot/pkg/rankingdata"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	Discord     *discordgo.Session
	RankingData *rankingdata.RankingData
}

func NewDiscordBot(token string, rankingPath string) (*DiscordBot, error) {
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	//discord.LogLevel = discordgo.LogInformational

	rankingDataPtr, err := rankingdata.ReadRankingData(rankingPath)
	if err != nil {
		return nil, err
	}

	return &DiscordBot{Discord: discord, RankingData: rankingDataPtr}, nil
}

func (bot *DiscordBot) Start() error {
	bot.Discord.AddHandler(handleMessageCreate)
	err := bot.Discord.Open()
	if err != nil {
		return err
	}
	return nil
}

func (bot *DiscordBot) Stop() {
	bot.Discord.Close()
}

func handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		//fmt.Println("Ignoring message from self")
		return
	}

	//dump some info about the message
	fmt.Println("Message received:")
	fmt.Println("  Author ID: " + m.Author.ID)
	fmt.Println("  Author Username: " + m.Author.Username)
	fmt.Println("  Guild: " + m.GuildID)
	fmt.Println("  Channel: " + m.ChannelID)

	fmt.Println("  Content: " + m.Content)
	fmt.Println("  m: " + m.ContentWithMentionsReplaced())
	fmt.Println("  Mentions:")
	for _, name := range m.Mentions {
		fmt.Println("    " + name.Username)
	}
	fmt.Println("  Mentions Everyone: " + fmt.Sprint(m.MentionEveryone))
	fmt.Println("  Mentions Roles:")
	for _, role := range m.MentionRoles {
		fmt.Println("    " + role)
	}
	fmt.Println("  Mentions Channels:")
	for _, channel := range m.MentionChannels {
		fmt.Println("    " + channel.Name)
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
		_, _ = s.ChannelMessageSend(m.ChannelID, "Command: "+command)
		switch command {
		case "help":
			_, _ = s.ChannelMessageSend(m.ChannelID, "Help is on the way!")
		case "register", "join", "add":
			_, _ = s.ChannelMessageSend(m.ChannelID, "Registering...")
		case "challenge":
			if len(m.Mentions) == 1 {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Challenging "+m.Mentions[0].Username+"...")
			} else {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Please mention one person to challenge.")
			}
		case "result":
			if len(words) >= 3 {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Please use one of: w, win, l, lose, f, forfeit.")
			} else {
				switch words[1] {
				case "w", "win":
					_, _ = s.ChannelMessageSend(m.ChannelID, "Recording win...")
				case "l", "lose":
					_, _ = s.ChannelMessageSend(m.ChannelID, "Recording loss...")
				case "f", "forfeit":
				default:
					_, _ = s.ChannelMessageSend(m.ChannelID, "Please use one of: w, win, l, lose, f, forfeit.")
				}
			}
		case "cancel":
			_, _ = s.ChannelMessageSend(m.ChannelID, "Cancelling...")
		case "ladder":
			_, _ = s.ChannelMessageSend(m.ChannelID, "Printing ladder...")
		case "history":
			_, _ = s.ChannelMessageSend(m.ChannelID, "Printing history...")
		case "active", "challenges":
			_, _ = s.ChannelMessageSend(m.ChannelID, "Printing active challenges...")
		case "set":
			_, _ = s.ChannelMessageSend(m.ChannelID, "Setting variable to value...")
		case "unset":
			_, _ = s.ChannelMessageSend(m.ChannelID, "Unsetting variable...")
		case "quit":
			_, _ = s.ChannelMessageSend(m.ChannelID, "Quitting...")
		case "ping":
			_, _ = s.ChannelMessageSend(m.ChannelID, "Pong!")
		case "pong":
			_, _ = s.ChannelMessageSend(m.ChannelID, "Ping!")
		default:
			_, _ = s.ChannelMessageSend(m.ChannelID, "Unknown command.")
		}
	}
}
