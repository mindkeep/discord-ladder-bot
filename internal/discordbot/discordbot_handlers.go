package discordbot

import (
	"discord_ladder_bot/internal/rankingdata"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func handleRegister(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	playerID := i.Member.User.ID
	gamename := ""
	for _, option := range o {
		switch option.Name {
		case "alt_user":
			if option.Type == discordgo.ApplicationCommandOptionUser {
				// this is optional, we user the user who sent the message if not specified
				if !c.IsAdmin(i.Member.User.ID) {
					return "You must be an admin to register other users.", nil
				}
				playerID = option.UserValue(nil).ID
			} else {
				return "", errors.New("internal error, unexpected option type, expected discord user")
			}

		case "gamename":
			gamename = option.StringValue()
		default:
			return "", errors.New("invalid option to register user: " + option.Name)
		}
	}
	if gamename == "" {
		gamename = "Unknown"
	}

	return c.AddPlayer(playerID, gamename)
}

func handleUnregister(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	playerID := i.Member.User.ID
	for _, option := range o {
		switch option.Name {
		case "alt_user":
			if option.Type != discordgo.ApplicationCommandOptionUser {
				return "", errors.New("internal error, unexpected option type, expected discord user")
			}
			if !c.IsAdmin(i.Member.User.ID) {
				return "You must be an admin to unregister other users.", nil
			}
			playerID = option.UserValue(nil).ID
		default:
			return "", errors.New("invalid option to unregister user: " + option.Name)
		}
	}
	return c.RemovePlayer(playerID)
}

func handleChallenge(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	challengerID := i.Member.User.ID
	defenderID := ""
	for _, option := range o {
		switch option.Name {
		case "alt_user":
			if option.Type != discordgo.ApplicationCommandOptionUser {
				return "", errors.New("internal error, unexpected option type, expected discord user")
			}
			if !c.IsAdmin(i.Member.User.ID) {
				return "You must be an admin to unregister other users.", nil
			}
			challengerID = option.UserValue(nil).ID
		case "defender":
			if option.Type != discordgo.ApplicationCommandOptionUser {
				return "", errors.New("internal error, unexpected option type, expected discord user")
			}
			defenderID = option.UserValue(nil).ID
		default:
			return "", errors.New("invalid option to challenge user: " + option.Name)
		}
	}
	if defenderID == "" {
		return "Please specify a defender to challenge.", nil
	}

	return c.StartChallenge(challengerID, defenderID)
}

func handleResult(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	result := ""
	playerID := i.Member.User.ID
	for _, option := range o {
		switch option.Name {
		case "alt_user":
			if option.Type != discordgo.ApplicationCommandOptionUser {
				return "", errors.New("internal error, unexpected option type, expected discord user")
			}
			if !c.IsAdmin(i.Member.User.ID) {
				return "You must be an admin to set results for other users.", nil
			}
			playerID = option.UserValue(nil).ID
		case "result":
			result = option.StringValue()
			if result != "won" && result != "lost" {
				return "Please specify a valid result (won, lost)", nil
			}
		default:
			return "", errors.New("invalid option to set challenge result: " + option.Name)
		}
	}

	return c.ResolveChallenge(playerID, result)
}

func handleCancel(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	playerID := i.Member.User.ID
	for _, option := range o {
		switch option.Name {
		case "alt_user":
			if option.Type != discordgo.ApplicationCommandOptionUser {
				return "", errors.New("internal error, unexpected option type, expected discord user")
			}
			if !c.IsAdmin(i.Member.User.ID) {
				return "You must be an admin to cancel challenges for other users.", nil
			}
			playerID = option.UserValue(nil).ID
		default:
			return "", errors.New("invalid option to cancel challenge: " + option.Name)
		}
	}
	return c.ResolveChallenge(playerID, "cancel")
}

func handleForfeit(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {
	playerID := i.Member.User.ID
	for _, option := range o {
		switch option.Name {
		case "alt_user":
			if option.Type != discordgo.ApplicationCommandOptionUser {
				return "", errors.New("internal error, unexpected option type, expected discord user")
			}
			if !c.IsAdmin(i.Member.User.ID) {
				return "You must be an admin to forfeit challenges for other users.", nil
			}
			playerID = option.UserValue(nil).ID
		default:
			return "", errors.New("invalid option to forfeit challenge: " + option.Name)
		}
	}
	return c.ResolveChallenge(playerID, "forfeit")
}

func handleUserSettings(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	playerID := i.Member.User.ID

	// if the user is an admin, they can set other users
	// find user first so that other settings apply to the right user
	for _, option := range o {
		if option.Name == "alt_user" && option.Type == discordgo.ApplicationCommandOptionUser {
			// this is optional, we user the user who sent the message if not specified
			if !c.IsAdmin(i.Member.User.ID) {
				return "You must be an admin to set other users.", nil
			}
			playerID = option.UserValue(nil).ID
		}
	}

	// loop through other options
	for _, option := range o {

		switch option.Name {
		case "alt_user":
			// already handled above
		case "status":
			err := c.SetPlayerStatus(playerID, option.StringValue())
			if err != nil {
				return "", err
			}
		case "gamename":
			err := c.SetPlayerGameName(playerID, option.StringValue())
			if err != nil {
				return "", err
			}
		case "notes":
			err := c.SetPlayerNotes(playerID, option.StringValue())
			if err != nil {
				return "", err
			}
		default:
			return "", fmt.Errorf("invalid option to set user settings: %s", option.Name)
		}
	}

	// get the updated settings
	player, err := c.FindPlayer(playerID)
	if err != nil {
		return "", err
	}

	// return the updated settings
	var response string
	response += fmt.Sprintf("User settings updated for <@%s>:\n", playerID)
	response += fmt.Sprintf("  gamename: %s\n", player.GameName)
	response += fmt.Sprintf("  status: %s\n", player.Status)
	response += fmt.Sprintf("  notes: %s\n", player.Notes)
	return response, nil
}

func handleSystemSettings(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	if !c.IsAdmin(i.Member.User.ID) {
		return "You must be an admin to set system settings.", nil
	}

	// loop through options
	// we don't need to check for user, since only admins can set system settings
	for _, option := range o {
		switch option.Name {
		case "mode":
			err := c.SetGameMode(option.StringValue())
			if err != nil {
				return "", err
			}
		case "timeoutdays":
			err := c.SetTimeout(int(option.IntValue()))
			if err != nil {
				return "", err
			}
		case "admin_add":
			err := c.AddAdmin(option.UserValue(nil).ID)
			if err != nil {
				return err.Error(), nil
			}
		case "admin_remove":
			err := c.RemoveAdmin(option.UserValue(nil).ID)
			if err != nil {
				return err.Error(), nil
			}
		case "notes":
			err := c.SetNotes(option.StringValue())
			if err != nil {
				return err.Error(), nil
			}
		default:
			return "", fmt.Errorf("invalid option to set system settings: %s", option.Name)
		}
	}
	var response string
	response += "Game settings:\n"
	response += fmt.Sprintf("  gamemode: %s\n", c.ChallengeMode)
	response += fmt.Sprintf("  timeout: %d (days)\n", c.ChallengeTimeoutDays)
	response += "  admins: "
	for _, admin := range c.Admins {
		response += fmt.Sprintf("<@%s> ", admin)
	}
	response += "\n"
	response += fmt.Sprintf("  notes: %s\n", c.Notes)
	return response, nil
}

func handleMove(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	playerID := ""
	position := -1

	if !c.IsAdmin(i.Member.User.ID) {
		return "You must be an admin to move players.", nil
	}

	error_response := "Please specify a player and a position."
	if len(o) != 2 {
		return error_response, nil
	}

	for _, option := range o {
		switch option.Name {
		case "user":
			if option.Type != discordgo.ApplicationCommandOptionUser {
				return "", errors.New("internal error, unexpected option type, expected discord user")
			}
			playerID = option.UserValue(nil).ID
		case "position":
			if option.Type != discordgo.ApplicationCommandOptionInteger {
				return "", errors.New("internal error, unexpected option type, expected integer")
			}
			position = int(option.IntValue())
		default:
			return "", fmt.Errorf("invalid option to move player: %s", option.Name)
		}
	}

	if position <= 0 {
		return "", errors.New("position must be greater than 0")
	}
	if playerID == "" {
		return "", errors.New("player not specified")
	}

	return c.MovePlayer(playerID, position)
}
