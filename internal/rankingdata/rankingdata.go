package rankingdata

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"discord_ladder_bot/internal/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RankingData struct {
	Version  string                `bson:"version,omitempty"`
	Channels []*ChannelRankingData `bson:"channels"`
	conf     *config.Config
	mutex    sync.Mutex
}

type ChannelRankingData struct {
	ChannelID            string          `bson:"channel_id"`
	ChallengeMode        string          `bson:"challenge_mode"`
	ChallengeTimeoutDays time.Duration   `bson:"challenge_timeout_days"`
	RankedPlayers        []Player        `bson:"ranked_players"`
	ActiveChallenges     []Challenge     `bson:"active_challenges"`
	ResultHistory        []ResultHistory `bson:"result_history"`
	Admins               []string        `bson:"admins"`
	mutex                sync.Mutex
}

type Player struct {
	PlayerID string `bson:"player_id"`
	Position int    `bson:"position"`
	Status   string `bson:"status,omitempty"`
	GameName string `bson:"game_name,omitempty"`
	Notes    string `bson:"notes,omitempty"`
}

type Challenge struct {
	ChallengerID      string    `bson:"challenger_id"`
	ChallengeeID      string    `bson:"challengee_id"`
	ChallengeDate     time.Time `bson:"challenge_date"`
	ChallengeDeadline time.Time `bson:"challenge_deadline"`
}

type ResultHistory struct {
	ChallengerID  string    `bson:"challenger_id"`
	ChallengeeID  string    `bson:"challengee_id"`
	Result        string    `bson:"result"`
	ChallengeDate time.Time `bson:"challenge_date,omitempty"`
	ResolveDate   time.Time `bson:"resolve_date,omitempty"`
}

// Locks the ranking data for a channel
func (c *ChannelRankingData) Lock() {
	c.mutex.Lock()
}

// Unlocks the ranking data for a channel
func (c *ChannelRankingData) Unlock() {
	c.mutex.Unlock()
}

// function that reads a mongodb and returns a RankingData struct
func ReadRankingData(conf *config.Config) (*RankingData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conf.MongoURI))
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(ctx)

	rankingData := RankingData{conf: conf}
	rankingData.Channels = make([]*ChannelRankingData, 0)

	db := client.Database(conf.MongoDBName)
	collection := db.Collection(conf.MongoCollectionName)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		fmt.Println("Warning: No ranking data found, creating new collection")
		// return an empty ranking data
		return &rankingData, nil
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		channelRankingData := ChannelRankingData{}
		err := cursor.Decode(&channelRankingData)
		if err != nil {
			return nil, err
		}
		rankingData.Channels = append(rankingData.Channels, &channelRankingData)
	}

	return &rankingData, nil
}

// function that writes a RankingData struct to a mongodb
func (rankingData *RankingData) Write() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(rankingData.conf.MongoURI))
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)

	collection := client.Database(rankingData.conf.MongoDBName).Collection(rankingData.conf.MongoCollectionName)
	// delete all documents in the collection, ignore errors for empty results
	_ = collection.Drop(ctx)

	// insert each channel into the collection
	for i := range rankingData.Channels {
		channel := &rankingData.Channels[i]
		_, err := collection.InsertOne(ctx, channel)
		if err != nil {
			return err
		}
	}

	return nil
}

//
// Utility/Private functions
// NOTE: These functions are not thread safe and should only be called from
// 	 within a function that has already locked the mutex.
//

// function that finds a channel in a RankingData struct
func (rankingData *RankingData) findChannel(channelID string) (*ChannelRankingData, error) {
	for i := range rankingData.Channels {
		channel := rankingData.Channels[i]
		if channel.ChannelID == channelID {
			return channel, nil
		}
	}
	return nil, errors.New("channel not found")
}

// function that finds a player in a RankingData channel struct
func (channel *ChannelRankingData) findPlayer(playerID string) (*Player, error) {
	for i := range channel.RankedPlayers {
		player := &channel.RankedPlayers[i]
		if player.PlayerID == playerID {
			return player, nil
		}
	}
	return nil, errors.New("player not found")
}

// function that finds a challenge in a RankingData channel struct
func (channel *ChannelRankingData) findChallenge(playerID string) (*Challenge, error) {
	for i := range channel.ActiveChallenges {
		challenge := &channel.ActiveChallenges[i]
		if challenge.ChallengerID == playerID || challenge.ChallengeeID == playerID {
			return challenge, nil
		}
	}
	return nil, errors.New("challenge not found")
}

// function that determines if a player is available for a challenge
func (channel *ChannelRankingData) isPlayerAvailable(playerID string) bool {
	_, err := channel.findChallenge(playerID)
	player, _ := channel.findPlayer(playerID)

	// return true if the player is not in a challenge and is active
	return err != nil && player.Status == "active"
}

// function that returns the tier of a position
func tierFromPos(position int) int {
	tier := 1
	tierdiv := 1
	for tierdiv < position {
		tier++
		tierdiv += tier
	}
	return tier
}

// function that returns the maximum position in a tier.
func maxPosInTier(tier int) int {
	t := 0
	pos := 0
	for t < tier {
		t++
		pos += t
	}
	return pos
}

// Position sorting functions
type byPosition []Player

func (a byPosition) Len() int           { return len(a) }
func (a byPosition) Less(i, j int) bool { return a[i].Position < a[j].Position }
func (a byPosition) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// private function that sorts the players by position and fixes any gaps
func (channel *ChannelRankingData) fixPositions() {

	// sort the players by position
	sort.Sort(byPosition(channel.RankedPlayers))

	// remove any gaps in the positions
	for i := range channel.RankedPlayers {
		channel.RankedPlayers[i].Position = i + 1
	}
}

//
// Public functions
//

func (channel *ChannelRankingData) FindPlayer(playerID string) (Player, error) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	player, err := channel.findPlayer(playerID)

	// return a copy of the player struct
	return *player, err
}

// function that sets the game mode for a channel
func (channel *ChannelRankingData) SetGameMode(gameMode string) error {

	//verify the game mode is valid
	if gameMode != "ladder" && gameMode != "pyramid" && gameMode != "linear" && gameMode != "open" {
		return errors.New("invalid game mode")
	}

	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	channel.ChallengeMode = gameMode
	return nil
}

// function tht sets the timeout of matches for a channel
func (channel *ChannelRankingData) SetTimeout(timeoutDays int) error {
	if timeoutDays < 1 || timeoutDays > 30 {
		return errors.New("timeout days must be between 1 and 30")
	}
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	channel.ChallengeTimeoutDays = time.Duration(timeoutDays) * 24 * time.Hour
	return nil
}

// function that adds an admin to a channel
func (channel *ChannelRankingData) AddAdmin(playerID string) error {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	// check if the player is already an admin
	for _, admin := range channel.Admins {
		if admin == playerID {
			return errors.New("player is already an admin")
		}
	}

	// add the player to the admin list
	channel.Admins = append(channel.Admins, playerID)
	return nil
}

// function that removes an admin from a channel
func (channel *ChannelRankingData) RemoveAdmin(playerID string) error {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	// make sure there's at least one admin left
	if len(channel.Admins) == 1 {
		return errors.New("cannot remove last admin")
	}

	// check if the player is an admin
	for i, admin := range channel.Admins {
		if admin == playerID {
			// remove the player from the admin list
			channel.Admins = append(channel.Admins[:i], channel.Admins[i+1:]...)
			return nil
		}
	}

	return errors.New("player is not an admin")
}

// function that prints a RankingData struct
func (channel *ChannelRankingData) PrintRaw() (string, error) {
	//lock the mutex
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	bsonBytes, err := bson.MarshalExtJSON(channel, false, false)
	if err != nil {
		return "", err
	}
	return string(bsonBytes), err
}

// function that returns a Discord formatted string of the ranking ladder
func (channel *ChannelRankingData) PrintRankings() (string, error) {
	// lock the mutex
	channel.mutex.Lock()
	defer channel.mutex.Unlock()
	tier := 0
	tierdiv := 1
	var response string

	// give a sane response if no one is registered
	if len(channel.RankedPlayers) == 0 {
		return "No players registered", nil
	}

	// walk through the players and build the string
	for i, player := range channel.RankedPlayers {
		// sanity check the position
		if player.Position != i+1 {
			return "", errors.New("internal player position is not correct")
		}

		pos := i + 1
		// print tier headings
		if channel.ChallengeMode == "pyramid" && pos == tierdiv {
			tier++
			response += fmt.Sprintf("## Tier %d:\n", tier)
			tierdiv += tier
		}

		chal, err := channel.findChallenge(player.PlayerID)
		if err != nil {
			// player is not in a challenge
			response += fmt.Sprintf("%d. %s/<@%s>\n", pos, player.GameName, player.PlayerID)
		} else {
			if chal.ChallengerID == player.PlayerID {
				// player is the challenger
				challengee, err := channel.findPlayer(chal.ChallengeeID)
				if err != nil {
					return "", errors.New("challengee not found")
				}
				response += fmt.Sprintf("%d. %s/<@%s> (challenging %s/<@%s>)\n", pos,
					player.GameName, player.PlayerID,
					challengee.GameName, chal.ChallengeeID)
			} else {
				// player is the challengee
				challenger, err := channel.findPlayer(chal.ChallengerID)
				if err != nil {
					return "", errors.New("challenger not found")
				}
				response += fmt.Sprintf("%d. %s/<@%s> (challenged by %s/<@%s>)\n", pos,
					player.GameName, player.PlayerID,
					challenger.GameName, chal.ChallengerID)
			}
		}
	}

	return response, nil
}

// function that verifies if a player is an admin
func (channel *ChannelRankingData) IsAdmin(playerID string) bool {
	//lock the mutex
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	// if the admin list is empty, then everyone is an admin
	if len(channel.Admins) == 0 {
		return true
	}

	// walk through the admins and see if the player is in the list
	for _, admin := range channel.Admins {
		if admin == playerID {
			return true
		}
	}
	return false
}

// function that returns a Discord formatted string of the active challenges
func (channel *ChannelRankingData) PrintChallenges() (string, error) {
	//lock the mutex
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	var response string
	// walk through the challenges and build the string
	for _, challenge := range channel.ActiveChallenges {
		challengee, err := channel.findPlayer(challenge.ChallengeeID)
		if err != nil {
			return "", errors.New("challengee not found")
		}
		challenger, err := channel.findPlayer(challenge.ChallengerID)
		if err != nil {
			return "", errors.New("challenger not found")
		}
		response += fmt.Sprintf("%s/<@%s>(#%d) vs %s/<@%s>(#%d)\n",
			challenger.GameName, challenger.PlayerID, challenger.Position,
			challengee.GameName, challengee.PlayerID, challengee.Position)
	}

	return response, nil
}

// function that returns a Discord formatted string of the result history
// TODO add result limit
func (channel *ChannelRankingData) PrintHistory() (string, error) {
	//lock the mutex
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	var response string
	// walk through the result history and build the string
	for _, result := range channel.ResultHistory {
		challengee, err := channel.findPlayer(result.ChallengeeID)
		if err != nil {
			return "", errors.New("challengee not found")
		}
		challenger, err := channel.findPlayer(result.ChallengerID)
		if err != nil {
			return "", errors.New("challenger not found")
		}
		// TODO add dates
		response += fmt.Sprintf("%s/<@%s> vs %s/<@%s> (%s)\n",
			challenger.GameName, challenger.PlayerID,
			challengee.GameName, challengee.PlayerID,
			result.Result)
	}

	return response, nil
}

// function that adds a new channel to the ranking data
func (rankingData *RankingData) AddChannel(channelID string, adminID string) (string, error) {
	rankingData.mutex.Lock()
	defer rankingData.mutex.Unlock()

	// return an error if the channel is already present
	if _, err := rankingData.findChannel(channelID); err == nil {
		return "", errors.New("channel is already registered")
	}

	// add the channel to the ranking data
	rankingData.Channels = append(rankingData.Channels,
		&ChannelRankingData{
			ChannelID:            channelID,
			ChallengeMode:        "ladder",
			ChallengeTimeoutDays: 7,
			RankedPlayers:        []Player{},
			ActiveChallenges:     []Challenge{},
			ResultHistory:        []ResultHistory{},
			Admins:               []string{adminID},
		})
	return "Let the games begin!", nil
}

// function that removes a channel from the ranking data
func (rankingData *RankingData) RemoveChannel(channelID string) (string, error) {
	rankingData.mutex.Lock()
	defer rankingData.mutex.Unlock()

	// return an error if the channel is not present
	if _, err := rankingData.findChannel(channelID); err != nil {
		return "", err
	}

	// remove the channel from the ranking data
	for i := range rankingData.Channels {
		if rankingData.Channels[i].ChannelID == channelID {
			rankingData.Channels = append(rankingData.Channels[:i], rankingData.Channels[i+1:]...)
			return fmt.Sprintf("Removed channel %s", channelID), nil
		}
	}
	return "", errors.New("channel not found")
}

// function that finds a channel in a RankingData struct
func (rankingData *RankingData) FindChannel(channelID string) (*ChannelRankingData, error) {
	rankingData.mutex.Lock()
	defer rankingData.mutex.Unlock()

	return rankingData.findChannel(channelID)
}

// function that adds a new player to the ranking data channel
func (channel *ChannelRankingData) AddPlayer(playerID string, gameName string) (string, error) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	// return an error if the player is already present
	if _, err := channel.findPlayer(playerID); err == nil {
		return "", errors.New("Player is already registered")
	}

	// add the player to the ranking data
	channel.RankedPlayers = append(channel.RankedPlayers,
		Player{
			PlayerID: playerID,
			Position: len(channel.RankedPlayers) + 1,
			GameName: gameName,
			Status:   "active",
			Notes:    "",
		})
	return fmt.Sprintf("Added %s/<@%s> to position %d",
		gameName,
		playerID,
		len(channel.RankedPlayers)), nil
}

// function that removes a player from the ranking data channel
func (channel *ChannelRankingData) RemovePlayer(playerID string) (string, error) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	// return an error if the player is not present
	removedPos := 0
	var gamename string
	for i := range channel.RankedPlayers {
		player := &channel.RankedPlayers[i]
		if player.PlayerID == playerID {
			removedPos = player.Position
			gamename = player.GameName
			// remove the player from the ranking data
			channel.RankedPlayers = append(channel.RankedPlayers[:i], channel.RankedPlayers[i+1:]...)
			break
		}
	}
	if removedPos == 0 {
		return "", errors.New("player not found")
	}

	// decrement the position of all players below the removed player
	channel.fixPositions()

	// remove any active challenges that the player is in
	for i := range channel.ActiveChallenges {
		challenge := &channel.ActiveChallenges[i]
		if challenge.ChallengerID == playerID || challenge.ChallengeeID == playerID {
			channel.ActiveChallenges = append(channel.ActiveChallenges[:i], channel.ActiveChallenges[i+1:]...)
			break
		}
	}

	return fmt.Sprintf("Removed %s/<@%s> from position %d",
		gamename,
		playerID,
		removedPos), nil
}

// function that moves a player to a new position
// NOTE: this function does not check if the move will cause invalid challenges
func (channel *ChannelRankingData) MovePlayer(playerID string, newPosition int) (string, error) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	maxPos := len(channel.RankedPlayers)

	if newPosition < 1 || newPosition > maxPos {
		return "", fmt.Errorf("invalid position, must be between 1 and %d", maxPos)
	}

	// return an error if the player is not present
	movingPlayer, err := channel.findPlayer(playerID)
	if err != nil {
		return "", errors.New("player not found")
	}
	gamename := movingPlayer.GameName

	// return error if the player is in a challenge
	if _, err := channel.findChallenge(playerID); err == nil {
		return "", errors.New("player is in a challenge")
	}

	for i := range channel.RankedPlayers {
		player := &channel.RankedPlayers[i]

		// if the player is moving up, decrement their position
		if player.Position <= newPosition && player.Position > movingPlayer.Position {
			player.Position--
		} else if player.Position >= newPosition {
			player.Position++
		}
	}
	movingPlayer.Position = newPosition

	// clean up the positions
	channel.fixPositions()

	return fmt.Sprintf("Moved %s/<@%s> to position %d",
		gamename,
		playerID,
		newPosition), nil
}

// function that starts a challenge
func (channel *ChannelRankingData) StartChallenge(challengerID string, challengeeID string) (string, error) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	// if the challenger is not registered, return an error
	challenger, err := channel.findPlayer(challengerID)
	if err != nil {
		return "", errors.New("challenger not found")
	}
	challengee, err := channel.findPlayer(challengeeID)
	if err != nil {
		return "", errors.New("challengee not found")
	}

	// if the challenger is not available, return an error
	// TODO: it would be good to make the reasoning for the error more specific
	if !channel.isPlayerAvailable(challengerID) {
		return "", errors.New("challenger is not available")
	}
	if !channel.isPlayerAvailable(challengeeID) {
		return "", errors.New("challengee is not available")
	}

	// if the challengee is a lower rank, it's invalid
	if challenger.Position < challengee.Position {
		return "", errors.New("challengee is a lower rank")
	}

	switch channel.ChallengeMode {
	// in linear/ladder mode, the challenger can only challenge the next person up
	case "linear", "ladder":
		if challenger.Position-1 != challengee.Position {
			return "", errors.New("challenger may only challenge the next person up")
		}
	// in pyramid mode, the challenger can only challenge someone in the same tier or the tier below
	case "pyramid":
		// determine if the challenger is eligible to challenge challengee
		challengerTier := tierFromPos(challenger.Position)
		challengeeTier := tierFromPos(challengee.Position)
		if challengerTier-challengeeTier > 1 {
			return "", errors.New("challenger must be within one tier of challengee")
		}
	case "open":
		// in open mode, the challenger can challenge anyone
	default:
		return "", errors.New("invalid challenge mode")
	}

	// create the challenge
	channel.ActiveChallenges = append(channel.ActiveChallenges,
		Challenge{
			ChallengerID:      challengerID,
			ChallengeeID:      challengeeID,
			ChallengeDate:     time.Now(),
			ChallengeDeadline: time.Now().Add(channel.ChallengeTimeoutDays),
		})
	response := fmt.Sprintf("Challenge started: %s/<@%s> vs %s/<@%s>",
		challenger.GameName, challenger.PlayerID,
		challengee.GameName, challengee.PlayerID)
	return response, nil
}

// TODO, have the cancel challenge function omit the challenge from the history

// function that resolves a challenge
func (channel *ChannelRankingData) ResolveChallenge(reporterID string, action string) (string, error) {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	var result string

	// find the challenge
	challenge, err := channel.findChallenge(reporterID)
	if err != nil {
		return "", errors.New("challenge not found")
	}

	// sanity check the action
	switch action {
	case "won", "lost", "cancel", "forfeit", "timed out":
		// do nothing
	default:
		return "", errors.New("invalid action")
	}

	// TODO this whole block is awkward since we're making challengee responsible
	// for reporting, just to flip the logic here.
	//
	// if the reporter is the challengee, reverse the result/action
	if reporterID == challenge.ChallengeeID {
		switch action {
		case "won":
			action = "lost"
		case "lost":
			action = "won"
		case "cancel":
			return "", errors.New("challengee cannot cancel, only forfeit")
		}
	} else if reporterID == challenge.ChallengerID {
		if action != "cancel" {
			return "", errors.New("challenger can only cancel")
		}
	} else {
		return "", errors.New("reporter is not in the challenge")
	}

	// add the result to the history only if not canceled
	if action != "cancel" {

		channel.ResultHistory = append(channel.ResultHistory,
			ResultHistory{
				ChallengerID:  challenge.ChallengerID,
				ChallengeeID:  challenge.ChallengeeID,
				Result:        action,
				ChallengeDate: challenge.ChallengeDate,
				ResolveDate:   time.Now(),
			})
	}

	challenger, err := channel.findPlayer(challenge.ChallengerID)
	if err != nil {
		return "", errors.New("challenger not found")
	}
	challengee, err := channel.findPlayer(challenge.ChallengeeID)
	if err != nil {
		return "", errors.New("challengee not found")
	}

	// if the challenger won (or the match was conceded or timed out), update the ranking
	if action == "won" || action == "forfeit" || action == "timed out" {
		challenger.Position, challengee.Position = challengee.Position, challenger.Position
		channel.fixPositions()
		result = fmt.Sprintf("Congratulations, %s/<@%s> has advanced from position %d to position %d!",
			challenger.GameName, challenger.PlayerID,
			challenger.Position, challengee.Position)
	} else if action == "lost" {
		result = fmt.Sprintf("Sorry, %s/<@%s>, better luck next time! %s/<@%s> holds position %d!",
			challengee.GameName, challengee.PlayerID,
			challenger.GameName, challenger.PlayerID,
			challenger.Position)
	} else if action == "cancel" {
		result = fmt.Sprintf("%s/<@%s> canceled challenge to %s/<@%s>",
			challenger.GameName, challenger.PlayerID,
			challengee.GameName, challengee.PlayerID)
	}

	// remove the challenge
	for i := range channel.ActiveChallenges {
		challenge := &channel.ActiveChallenges[i]
		if challenge.ChallengerID == reporterID || challenge.ChallengeeID == reporterID {
			channel.ActiveChallenges = append(channel.ActiveChallenges[:i], channel.ActiveChallenges[i+1:]...)
			break
		}
	}

	return result, nil
}

// function that sets a player's availability
func (channel *ChannelRankingData) SetPlayerStatus(playerID string, status string) error {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	// return an error if the player is not present
	player, err := channel.findPlayer(playerID)
	if err != nil {
		return errors.New("player not found")
	}

	player.Status = status
	return nil
}

// function that sets a player's gamename
func (channel *ChannelRankingData) SetPlayerGameName(playerID string, gameName string) error {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	// return an error if the player is not present
	player, err := channel.findPlayer(playerID)
	if err != nil {
		return errors.New("player not found")
	}

	player.GameName = gameName
	return nil
}

// function that sets a player's notes
func (channel *ChannelRankingData) SetPlayerNotes(playerID string, notes string) error {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	// return an error if the player is not present
	player, err := channel.findPlayer(playerID)
	if err != nil {
		return errors.New("player not found")
	}

	player.Notes = notes
	return nil
}
