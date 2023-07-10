package discordbot

import (
	"discord_ladder_bot/internal/rankingdata"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func handleRegister(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	playerID := i.Member.User.ID
	for _, option := range o {
		if option.Name == "user" && option.Type == discordgo.ApplicationCommandOptionUser {
			// this is optional, we user the user who sent the message if not specified
			if !c.IsAdmin(i.Member.User.ID) {
				return "You must be an admin to register other users.", nil
			}

			playerID = option.UserValue(nil).ID
		} else {
			return "", nil
		}
	}

	err := c.AddPlayer(playerID)
	if err != nil {
		return "", err
	}
	return "Registered!", nil
}

func handleUnregister(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	playerID := i.Member.User.ID
	for _, option := range o {
		if option.Name == "user" && option.Type == discordgo.ApplicationCommandOptionUser {
			// this is optional, we user the user who sent the message if not specified
			if !c.IsAdmin(i.Member.User.ID) {
				return "You must be an admin to register other users.", nil
			}
			playerID = option.UserValue(nil).ID
		} else {
			return "", nil
		}
	}
	err := c.RemovePlayer(playerID)
	if err != nil {
		return "", err
	}
	return "Unregistered!", nil
}

func handleChallenge(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	if len(o) != 1 || o[0].Type != discordgo.ApplicationCommandOptionUser {
		return "Please specify a player to challenge", nil
	}
	playerID := o[0].UserValue(nil).ID

	err := c.StartChallenge(i.Member.User.ID, playerID)
	if err != nil {
		return "", err
	}

	return "Challenge started!", nil
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

func handleSet(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	error_response := "Please specify user or system"

	if len(o) != 1 || o[0].Type != discordgo.ApplicationCommandOptionSubCommand {
		return error_response, nil
	}
	switch o[0].Name {
	case "user":
		return handleSetUser(c, i, o[0].Options)
	case "system":
		return handleSetSystem(c, i, o[0].Options)
	default:
		return error_response, nil
	}
}

func handleSetUser(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	playerID := i.Member.User.ID
	var key, value string
	for _, option := range o {
		if option.Name == "user" && option.Type == discordgo.ApplicationCommandOptionUser {
			// this is optional, we user the user who sent the message if not specified
			if !c.IsAdmin(i.Member.User.ID) {
				return "You must be an admin to set other users.", nil
			}
			playerID = option.UserValue(nil).ID
		} else if option.Name == "key" && option.Type == discordgo.ApplicationCommandOptionString {
			key = option.StringValue()
		} else if option.Name == "value" && option.Type == discordgo.ApplicationCommandOptionString {
			value = option.StringValue()
		} else {
			return "", errors.New("invalid option to set user")
		}
	}

	switch key {
	case "status":
		if value != "active" && value != "inactive" {
			return "Invalid status, must be active or inactive", nil
		}
		c.SetPlayerStatus(playerID, value)
		return "Status set to " + value + "!", nil
	case "notes":
		if len(value) > 100 {
			return "Notes must be less than 100 characters", nil
		}
		c.SetPlayerNotes(playerID, value)
		return "Notes set to " + value + "!", nil
	default:
		return "Invalid key, must be status or notes", nil
	}
}

func handleSetSystem(c *rankingdata.ChannelRankingData,
	i *discordgo.InteractionCreate,
	o []*discordgo.ApplicationCommandInteractionDataOption) (string, error) {

	if len(o) == 0 {
		var response string
		response += "Game settings:\n"
		response += fmt.Sprintf("  ChallengeMode: %s (ladder or pyramid or open)\n", c.ChallengeMode)
		response += fmt.Sprintf("  ChallengeTimeoutDays: %d\n", c.ChallengeTimeoutDays)
		response += fmt.Sprintf("  Admins: %s\n", strings.Join(c.Admins, ", "))
		return response, nil
	}

	if !c.IsAdmin(i.Member.User.ID) {
		return "You must be an admin to set system settings.", nil
	}

	var key, value string
	for _, option := range o {

		if option.Name == "key" && option.Type == discordgo.ApplicationCommandOptionString {
			key = option.StringValue()
		} else if option.Name == "value" && option.Type == discordgo.ApplicationCommandOptionString {
			value = option.StringValue()
		} else {
			return "", errors.New("invalid option to set system")
		}
	}

	switch key {
	case "mode":
		if value != "ladder" && value != "pyramid" && value != "open" {
			return "Invalid ChallengeMode, must be ladder, pyramid, or open", nil
		}
		c.ChallengeMode = value
		return "ChallengeMode set to " + value + "!", nil
	case "timeoutdays":
		timeout, err := strconv.Atoi(value)
		if err != nil {
			return "Invalid ChallengeTimeoutDays, must be an integer", nil
		}
		c.ChallengeTimeoutDays = time.Duration(timeout) * 24 * time.Hour
		return "ChallengeTimeoutDays set to " + value + "!", nil
	case "admin_add":
		if !c.IsAdmin(value) {
			c.Admins = append(c.Admins, value)
			return value + " added to admins!", nil
		} else {
			return value + " is already an admin!", nil
		}
	case "admin_remove":
		if c.IsAdmin(value) {
			for i, admin := range c.Admins {
				if admin == value {
					c.Admins = append(c.Admins[:i], c.Admins[i+1:]...)
					return value + " removed from admins!", nil
				}
			}
		} else {
			return value + " is not an admin!", nil
		}
	default:
		return "Invalid key, must be mode, timeoutdays, admin_add, or admin_remove", nil
	}

	return "", errors.New("um, this shouldn't happen")
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

	err := c.MovePlayer(userID, position)
	if err != nil {
		return "", err
	}
	return "Player moved!", nil
}
