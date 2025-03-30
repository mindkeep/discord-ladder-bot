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
	gamename := i.Member.User.Username
	for _, option := range o {
		switch option.Name {
		case "user":
			if option.Type == discordgo.ApplicationCommandOptionUser {
				// this is optional, we user the user who sent the message if not specified
				if !c.IsAdmin(i.Member.User.ID) {
					return "You must be an admin to register other users.", nil
				}

				playerID = option.UserValue(nil).ID
				gamename = option.UserValue(nil).Username
			} else {
				return "", errors.New("internal error, unexpected option type, expected discord user")
			}

		case "gamename":
			if len(option.StringValue()) > 100 {
				return "gamename must be less than 100 characters", nil
			}
			gamename = option.StringValue()
		default:
			return "", errors.New("invalid option to register user: " + option.Name)
		}
	}

	return c.AddPlayer(playerID, gamename)
}

func handleUnregister(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	playerID := i.Member.User.ID
	for _, option := range o {
		if option.Name == "user" && option.Type == discordgo.ApplicationCommandOptionUser {
			// this is optional, we user the user who sent the message if not specified
			if !c.IsAdmin(i.Member.User.ID) {
				return "You must be an admin to unregister other users.", nil
			}
			playerID = option.UserValue(nil).ID
		} else {
			return "", nil
		}
	}
	return c.RemovePlayer(playerID)
}

func handleChallenge(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	if len(o) != 1 || o[0].Type != discordgo.ApplicationCommandOptionUser {
		return "Please specify a player to challenge", nil
	}
	playerID := o[0].UserValue(nil).ID

	response, err := c.StartChallenge(i.Member.User.ID, playerID)
	if err != nil {
		return "", err
	}

	return response, nil
}

func handleResult(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	if len(o) != 1 || o[0].Type != discordgo.ApplicationCommandOptionString {
		return "Please specify a result (w, won, l, lost)", nil
	}
	result := o[0].StringValue()

	// we tried to be specific in the command, but people will still mess it up
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
	response, err := c.ResolveChallenge(i.Member.User.ID, result)
	if err != nil {
		return "", err
	}
	return response, nil
}

func handleCancel(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	response, err := c.ResolveChallenge(i.Member.User.ID, "cancel")
	if err != nil {
		return "", err
	}
	return response, nil
}

func handleForfeit(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	response, err := c.ResolveChallenge(i.Member.User.ID, "forfeit")
	if err != nil {
		return "", err
	}
	return response, nil
}

func handleUserSettings(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	playerID := i.Member.User.ID

	// if the user is an admin, they can set other users
	// find user first so that other settings apply to the right user
	for _, option := range o {
		if option.Name == "user" && option.Type == discordgo.ApplicationCommandOptionUser {
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
		case "user":
			// already handled above
		case "status":
			err := c.SetPlayerStatus(playerID, option.StringValue())
			if err != nil {
				return "", err
			}
		case "gamename":
			if len(option.StringValue()) > 100 {
				return "Game name must be less than 100 characters", nil
			}
			err := c.SetPlayerGameName(playerID, option.StringValue())
			if err != nil {
				return "", err
			}
		case "notes":
			if len(option.StringValue()) > 100 {
				return "notes must be less than 100 characters", nil
			}
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

	userID := i.Member.User.ID
	position := -1

	if !c.IsAdmin(userID) {
		return "You must be an admin to move players.", nil
	}

	error_response := "Please specify a player and a position."
	if len(o) != 2 {
		return error_response, nil
	}

	for _, option := range o {
		if option.Name == "user" && option.Type == discordgo.ApplicationCommandOptionUser {
			// this is optional, we user the user who sent the message if not specified
			userID = option.UserValue(nil).ID
		} else if option.Name == "position" && option.Type == discordgo.ApplicationCommandOptionInteger {
			position = int(option.IntValue())
		} else {
			return error_response, nil
		}
	}

	if position < 1 {
		return "Position must be greater than 0.", nil
	}

	return c.MovePlayer(userID, position)
}
