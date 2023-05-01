package rankingdata

import (
	"errors"
	"io"
	"os"
	"sort"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type RankingData struct {
	Version  string               `yaml:"version"`
	Channels []ChannelRankingData `yaml:"channels"`
	mutex    sync.Mutex
}

type ChannelRankingData struct {
	ChannelID        string          `yaml:"channel_id"`
	ChallengeMode    string          `yaml:"challenge_mode"`
	RankedPlayers    []Player        `yaml:"ranked_players"`
	ActiveChallenges []Challenge     `yaml:"active_challenges"`
	ResultHistory    []ResultHistory `yaml:"result_history"`
	mutex            sync.Mutex
}

type Player struct {
	PlayerID        string `yaml:"player_id"`
	Status          string `yaml:"status"`
	Position        int    `yaml:"position"`
	TimeZone        string `yaml:"time_zone,optional"`
	PrefferedServer string `yaml:"preffered_server,optional"`
}

type Challenge struct {
	ChallengerID      string    `yaml:"challenger_id"`
	ChallengeeID      string    `yaml:"challengee_id"`
	ChallengeDate     time.Time `yaml:"challenge_date"`
	ChallengeDeadline time.Time `yaml:"challenge_deadline"`
}

type ResultHistory struct {
	Challenger    string    `yaml:"challenger"`
	Challengee    string    `yaml:"challengee"`
	Result        string    `yaml:"result"`
	ChallengeDate time.Time `yaml:"challenge_date"`
	ResolveDate   time.Time `yaml:"resolve_date"`
}

// function that reads a json file and returns a RankingData struct
func ReadRankingData(path string) (*RankingData, error) {
	yamlFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer yamlFile.Close()

	yamlBytes, _ := io.ReadAll(yamlFile)
	var rankingData RankingData
	err = yaml.Unmarshal(yamlBytes, &rankingData)
	if err != nil {
		return nil, err
	}

	return &rankingData, nil
}

// function that writes a RankingData struct to a json file
func (rankingData *RankingData) Write(path string) error {

	//lock all the mutexes
	rankingData.mutex.Lock()
	defer rankingData.mutex.Unlock()
	for i := range rankingData.Channels {
		rankingData.Channels[i].mutex.Lock()
		defer rankingData.Channels[i].mutex.Unlock()
	}

	yamlBytes, err := yaml.Marshal(rankingData)
	if err != nil {
		return err
	}

	yamlFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer yamlFile.Close()

	_, err = yamlFile.Write(yamlBytes)
	if err != nil {
		return err
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
		channel := &rankingData.Channels[i]
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

func tierFromPos(position int) int {
	tier := 1
	tierdiv := 1
	for tierdiv < position {
		tier++
		tierdiv += tier
	}
	return tier
}

// maxPosInTier returns the maximum position in a tier.
func maxPosInTier(tier int) int {
	t := 1
	pos := 1
	for t < tier {
		pos += t
		t++
	}
	return pos
}

type byPosition []Player

func (a byPosition) Len() int           { return len(a) }
func (a byPosition) Less(i, j int) bool { return a[i].Position < a[j].Position }
func (a byPosition) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

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

// function that prints a RankingData struct
func (rankingData *RankingData) PrintRaw() (string, error) {
	//lock all the mutexes
	rankingData.mutex.Lock()
	defer rankingData.mutex.Unlock()

	for i := range rankingData.Channels {
		channel := &rankingData.Channels[i]
		channel.mutex.Lock()
		defer channel.mutex.Unlock()
	}

	yamlBytes, err := yaml.Marshal(rankingData)
	if err != nil {
		return "", err
	}
	return string(yamlBytes), err
}

// function that adds a new channel to the ranking data
func (rankingData *RankingData) AddChannel(channelID string) error {
	rankingData.mutex.Lock()
	defer rankingData.mutex.Unlock()

	// return an error if the channel is already present
	if _, err := rankingData.findChannel(channelID); err == nil {
		return errors.New("channel is already registered")
	}

	// add the channel to the ranking data
	rankingData.Channels = append(rankingData.Channels,
		ChannelRankingData{
			ChannelID:        channelID,
			ChallengeMode:    "ladder",
			RankedPlayers:    []Player{},
			ActiveChallenges: []Challenge{},
			ResultHistory:    []ResultHistory{},
		})
	return nil
}

// function that adds a new player to the ranking data channel
func (channel *ChannelRankingData) AddPlayer(playerID string) error {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	// return an error if the player is already present
	if _, err := channel.findPlayer(playerID); err == nil {
		return errors.New("Player is already registered")
	}

	// add the player to the ranking data
	channel.RankedPlayers = append(channel.RankedPlayers,
		Player{
			PlayerID: playerID,
			Status:   "active",
			Position: len(channel.RankedPlayers) + 1,
		})
	return nil
}

// function that removes a player from the ranking data channel
func (channel *ChannelRankingData) RemovePlayer(playerID string) error {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	// return an error if the player is not present
	removedPos := 0
	for i := range channel.RankedPlayers {
		player := &channel.RankedPlayers[i]
		if player.PlayerID == playerID {
			removedPos = player.Position
			channel.RankedPlayers = append(channel.RankedPlayers[:i], channel.RankedPlayers[i+1:]...)
			break
		}
	}
	if removedPos == 0 {
		return errors.New("player not found")
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

	return nil
}

// function that moves a player to a new position
// NOTE: this function does not check if the move will cause invalid challenges
func (channel *ChannelRankingData) MovePlayer(playerID string, newPosition int) error {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	// return an error if the player is not present
	movingPlayer, err := channel.findPlayer(playerID)
	if err != nil {
		return errors.New("player not found")
	}

	// return error if the player is in a challenge
	if _, err := channel.findChallenge(playerID); err == nil {
		return errors.New("player is in a challenge")
	}

	for i := range channel.RankedPlayers {
		player := &channel.RankedPlayers[i]
		if player.Position >= newPosition {
			player.Position++
		}
	}
	movingPlayer.Position = newPosition

	return nil
}

// function that starts a challenge
func (channel *ChannelRankingData) StartChallenge(channelID string, challengerID string, challengeeID string) error {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	// if the challenger is not registered, return an error
	challenger, err := channel.findPlayer(challengerID)
	if err != nil {
		return errors.New("challenger not found")
	}
	challengee, err := channel.findPlayer(challengeeID)
	if err != nil {
		return errors.New("challengee not found")
	}

	// if the challenger is not available, return an error
	// TODO: it would be good to make the reasoning for the error more specific
	if !channel.isPlayerAvailable(challengerID) {
		return errors.New("challenger is not available")
	}
	if !channel.isPlayerAvailable(challengeeID) {
		return errors.New("challengee is not available")
	}

	// determine if the challenger is eligible to challenge challengee
	challengerTier := tierFromPos(challenger.Position)
	challengeeTier := tierFromPos(challengee.Position)

	switch channel.ChallengeMode {
	// in linear/ladder mode, the challenger can only challenge the next person up
	case "linear", "ladder":
		if challenger.Position-1 != challengee.Position {
			return errors.New("challenger may only challenge the next person up")
		}
	// in pyramid mode, the challenger can only challenge someone in the same tier or the tier below
	case "pyramid":
		if challenger.Position > challengee.Position && challengerTier-1 <= challengeeTier {
			return errors.New("challenger is not eligible to challenge challengee")
		}
	default:
		return errors.New("invalid challenge mode")
	}

	// create the challenge
	channel.ActiveChallenges = append(channel.ActiveChallenges,
		Challenge{
			ChallengerID:      challengerID,
			ChallengeeID:      challengeeID,
			ChallengeDate:     time.Now(),
			ChallengeDeadline: time.Now().Add(time.Hour * 24 * 7),
		})
	return nil
}

// function that resolves a challenge
func (channel *ChannelRankingData) ResolveChallenge(reporterID string, action string) error {
	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	// find the challenge
	challenge, err := channel.findChallenge(reporterID)
	if err != nil {
		return errors.New("challenge not found")
	}

	// sanity check the action
	switch action {
	case "won", "lost", "cancel", "forfiet", "timed out":
		// do nothing
	default:
		return errors.New("invalid action")
	}

	// if the reporter is the challengee, reverse the result/action
	if reporterID == challenge.ChallengeeID {
		switch action {
		case "won":
			action = "lost"
		case "lost":
			action = "won"
		case "cancel":
			return errors.New("challengee cannot cancel, only forfiet")
		}
	} else if reporterID == challenge.ChallengerID {
		if action == "forfiet" {
			return errors.New("challenger cannot forfiet, only cancel")
		}
	}

	// add the result to the history
	channel.ResultHistory = append(channel.ResultHistory,
		ResultHistory{
			Challenger:    challenge.ChallengerID,
			Challengee:    challenge.ChallengeeID,
			Result:        action,
			ChallengeDate: challenge.ChallengeDate,
			ResolveDate:   time.Now(),
		})

	// if the challenger won (or the match was conceded or timed out), update the ranking
	if action == "won" || action == "forfiet" || action == "timed out" {
		challenger, err := channel.findPlayer(challenge.ChallengerID)
		if err != nil {
			return errors.New("challenger not found")
		}
		challengee, err := channel.findPlayer(challenge.ChallengeeID)
		if err != nil {
			return errors.New("challengee not found")
		}
		challenger.Position, challengee.Position = challengee.Position, challenger.Position
		channel.fixPositions()
	}

	// remove the challenge
	for i := range channel.ActiveChallenges {
		challenge := &channel.ActiveChallenges[i]
		if challenge.ChallengerID == reporterID || challenge.ChallengeeID == reporterID {
			channel.ActiveChallenges = append(channel.ActiveChallenges[:i], channel.ActiveChallenges[i+1:]...)
			break
		}
	}

	return nil
}
