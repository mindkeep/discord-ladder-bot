package discordbot

import (
	"discord_ladder_bot/pkg/rankingdata"
	"fmt"

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
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "ping" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Pong!")
	} else if m.Content == "print" {
		//_, _ = s.ChannelFileSend(m.ChannelID, "ranking.yml", "ranking.yml")
	} else {
		fmt.Println(m.Author.Username + " says: " + m.Content)
	}

}
