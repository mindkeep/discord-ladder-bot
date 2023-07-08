package discordbot

import (
	"discord_ladder_bot/internal/rankingdata"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func handle_register(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
	var playerID string
	var playerName string
	if len(m.Mentions) > 1 {
		return "You can only register one user at a time.", nil
	} else if len(m.Mentions) == 1 {
		if c.IsAdmin(m.Message.Author.ID) {
			playerID = m.Mentions[0].ID
			playerName = m.Mentions[0].Username
		} else {
			return "You must be an admin to register other users.", nil
		}
	} else {
		playerID = m.Message.Author.ID
		playerName = m.Message.Author.Username
	}
	err := c.AddPlayer(playerID, playerName)
	if err != nil {
		return "", err
	}
	return "Registered!", nil
}

func handle_unregister(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
	var playerID string
	if len(m.Mentions) > 1 {
		return "You can only unregister one user at a time.", nil
	} else if len(m.Mentions) == 1 {
		if c.IsAdmin(m.Message.Author.ID) {
			playerID = m.Mentions[0].ID
		} else {
			return "You must be an admin to unregister other users.", nil
		}
	} else {
		playerID = m.Message.Author.ID
	}
	err := c.RemovePlayer(playerID)
	if err != nil {
		return "", err
	}
	return "Unregistered!", nil
}

func handle_challenge(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
	if len(m.Mentions) != 1 {
		return "Please @ mention one person to challenge.", nil
	} else {
		err := c.StartChallenge(m.Author.ID, m.Mentions[0].ID)
		if err != nil {
			return "", err
		}
	}
	return "Challenge started!", nil
}

func handle_result(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
	words := strings.Split(m.Content, " ")
	if len(words) != 2 {
		return "Please use one of: w, won, l, lost", nil
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
	}
	err := c.ResolveChallenge(m.Author.ID, result)
	if err != nil {
		return "", err
	}
	return "Challenge has been resolved... somehow, TODO, add something clever...", nil
}

func handle_cancel(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
	err := c.ResolveChallenge(m.Author.ID, "cancel")
	if err != nil {
		return "", err
	}
	return "Challenge canceled!", nil
}

func handle_forfeit(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
	err := c.ResolveChallenge(m.Author.ID, "forfeit")
	if err != nil {
		return "", err
	}
	return "Challenge forfeited!", nil
}

func handle_set(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
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
}

func handle_move(c *rankingdata.ChannelRankingData, m *discordgo.MessageCreate) (string, error) {
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
}
