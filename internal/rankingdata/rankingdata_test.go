package rankingdata

import (
	"testing"
	"time"

	"github.com/magiconair/properties/assert"
)

func TestTierFromPos(t *testing.T) {
	testCases := []struct {
		pos      int
		expected int
	}{
		{1, 1},
		{2, 2},
		{3, 2},
		{4, 3},
		{5, 3},
		{6, 3},
		{7, 4},
		{8, 4},
		{9, 4},
		{10, 4},
		{11, 5},
		{12, 5},
		{13, 5},
		{14, 5},
		{15, 5},
		{16, 6},
		{17, 6},
		{18, 6},
		{19, 6},
		{20, 6},
		{21, 6},
		{22, 7},
		{23, 7},
		{29, 8},
		{36, 8},
		{37, 9},
		{45, 9},
		{46, 10},
		{55, 10},
		{56, 11},
	}
	for test := range testCases {
		assert.Equal(t, tierFromPos(testCases[test].pos), testCases[test].expected)
	}
}

func TestMaxPosInTier(t *testing.T) {
	testCases := []struct {
		tier     int
		expected int
	}{
		{1, 1},
		{2, 3},
		{3, 6},
		{4, 10},
		{5, 15},
		{6, 21},
		{7, 28},
		{8, 36},
		{9, 45},
		{10, 55},
		{11, 66},
		{12, 78},
		{13, 91},
		{14, 105},
		{15, 120},
	}
	for test := range testCases {
		assert.Equal(t, maxPosInTier(testCases[test].tier), testCases[test].expected)
	}
}

func TestAddPlayer(t *testing.T) {

	data := RankingData{
		Version: "v1_test",
		Channels: []*ChannelRankingData{
			{ChannelID: "1234", RankedPlayers: []Player{
				{PlayerID: "1234", GameName: "u1234", Status: "active", Position: 1},
				{PlayerID: "5678", GameName: "u5678", Status: "active", Position: 2},
				{PlayerID: "9012", GameName: "u9012", Status: "active", Position: 3},
				{PlayerID: "3456", GameName: "u3456", Status: "active", Position: 4},
				{PlayerID: "7890", GameName: "u7890", Status: "active", Position: 5},
			}}}}

	channel, err := data.findChannel("1234")
	if err != nil {
		t.Errorf("Error finding channel: %s", err)
		return
	}

	if _, err := channel.AddPlayer("1111", "u1111"); err != nil {
		t.Errorf("Error adding player: %s", err)
	}

	// check that the player was added
	if player, err := channel.findPlayer("1111"); err != nil {
		t.Errorf("Error finding player: %s", err)
	} else {

		assert.Equal(t, player.PlayerID, "1111")
		assert.Equal(t, player.GameName, "u1111")
		assert.Equal(t, player.Status, "active")
		assert.Equal(t, player.Position, 6)

	}
}

func TestRemovePlayer(t *testing.T) {
	data := RankingData{
		Version: "v1_test",
		Channels: []*ChannelRankingData{
			{ChannelID: "1234", RankedPlayers: []Player{
				{PlayerID: "1234", GameName: "u1234", Status: "active", Position: 1},
				{PlayerID: "5678", GameName: "u5678", Status: "active", Position: 2},
				{PlayerID: "9012", GameName: "u9012", Status: "active", Position: 3},
				{PlayerID: "3456", GameName: "u3456", Status: "active", Position: 4},
				{PlayerID: "7890", GameName: "u7890", Status: "active", Position: 5},
			}}}}

	channel, err := data.findChannel("1234")
	if err != nil {
		t.Errorf("Error finding channel: %s", err)
		return
	}

	// check that the player exists
	if _, err := channel.findPlayer("5678"); err != nil {
		t.Errorf("Error finding player: %s", err)
	}
	assert.Equal(t, len(channel.RankedPlayers), 5)

	if _, err := channel.RemovePlayer("5678"); err != nil {
		t.Errorf("Error removing player: %s", err)
	}

	// check that the player was removed
	if _, err := channel.findPlayer("5678"); err == nil {
		t.Errorf("Player was not removed")
	}
	assert.Equal(t, len(channel.RankedPlayers), 4)

	// check that the player positions were updated
	for i := range channel.RankedPlayers {
		assert.Equal(t, channel.RankedPlayers[i].Position, i+1)
	}

	// attempt to remove a player that doesn't exist
	if _, err := channel.RemovePlayer("1111"); err == nil {
		t.Errorf("Error removing player: %s", err)
	}
	assert.Equal(t, len(channel.RankedPlayers), 4)
}

func TestMovePlayer(t *testing.T) {
	data := RankingData{
		Version: "v1_test",
		Channels: []*ChannelRankingData{
			{ChannelID: "1234", RankedPlayers: []Player{
				{PlayerID: "1234", GameName: "u1234", Status: "active", Position: 1},
				{PlayerID: "5678", GameName: "u5678", Status: "active", Position: 2},
				{PlayerID: "9012", GameName: "u9012", Status: "active", Position: 3},
				{PlayerID: "3456", GameName: "u3456", Status: "active", Position: 4},
				{PlayerID: "7890", GameName: "u7890", Status: "active", Position: 5},
			}}}}

	channel, err := data.findChannel("1234")
	if err != nil {
		t.Errorf("Error finding channel: %s", err)
		return
	}

	// check that the player exists
	if _, err := channel.findPlayer("5678"); err != nil {
		t.Errorf("Error finding player: %s", err)
	}

	if _, err := channel.MovePlayer("5678", 3); err != nil {
		t.Errorf("Error moving player: %s", err)
	}

	// check that the player was moved
	if player, err := channel.findPlayer("5678"); err != nil {
		t.Errorf("Error finding player: %s", err)
	} else {
		assert.Equal(t, player.Position, 3)
	}

	// check that the player positions were updated
	for i := range channel.RankedPlayers {
		assert.Equal(t, channel.RankedPlayers[i].Position, i+1)
	}

	// attempt to move a player that doesn't exist
	if _, err := channel.MovePlayer("1111", 3); err == nil {
		t.Errorf("Error moving player: %s", err)
	}
}

func TestCleanupExpiredChallenges(t *testing.T) {
	data := RankingData{
		Version: "v1_test",
		Channels: []*ChannelRankingData{
			{
				ChannelID:            "1234",
				ChallengeTimeoutDays: 7 * 24 * time.Hour,
				RankedPlayers: []Player{
					{PlayerID: "1234", GameName: "u1234", Status: "active", Position: 1},
					{PlayerID: "5678", GameName: "u5678", Status: "active", Position: 2},
					{PlayerID: "9012", GameName: "u9012", Status: "active", Position: 3},
				},
				ActiveChallenges: []Challenge{
					{
						ChallengerID:      "5678",
						DefenderID:        "1234",
						ChallengeDate:     time.Now().Add(-10 * 24 * time.Hour),
						ChallengeDeadline: time.Now().Add(-3 * 24 * time.Hour), // Expired 3 days ago
					},
					{
						ChallengerID:      "9012",
						DefenderID:        "5678",
						ChallengeDate:     time.Now(),
						ChallengeDeadline: time.Now().Add(5 * 24 * time.Hour), // Still active
					},
				},
				ResultHistory: []ResultHistory{},
			},
		},
	}

	channel, err := data.findChannel("1234")
	if err != nil {
		t.Errorf("Error finding channel: %s", err)
		return
	}

	// Clean up expired challenges
	expiredCount, err := channel.CleanupExpiredChallenges()
	if err != nil {
		t.Errorf("Error cleaning up expired challenges: %s", err)
	}

	// Check that one challenge was expired
	assert.Equal(t, expiredCount, 1)

	// Check that one challenge remains active
	assert.Equal(t, len(channel.ActiveChallenges), 1)
	assert.Equal(t, channel.ActiveChallenges[0].ChallengerID, "9012")

	// Check that the expired challenge was added to result history
	assert.Equal(t, len(channel.ResultHistory), 1)
	assert.Equal(t, channel.ResultHistory[0].Result, "timed out")

	// Check that positions were swapped (challenger 5678 advanced from 2 to 1)
	player1, _ := channel.findPlayer("1234")
	player2, _ := channel.findPlayer("5678")
	assert.Equal(t, player2.Position, 1) // Challenger won and moved to position 1
	assert.Equal(t, player1.Position, 2) // Defender dropped to position 2
}

func TestRemovePlayerIfNotInGuild(t *testing.T) {
	data := RankingData{
		Version: "v1_test",
		Channels: []*ChannelRankingData{
			{
				ChannelID: "1234",
				RankedPlayers: []Player{
					{PlayerID: "1234", GameName: "u1234", Status: "active", Position: 1},
					{PlayerID: "5678", GameName: "u5678", Status: "active", Position: 2},
					{PlayerID: "9012", GameName: "u9012", Status: "active", Position: 3},
				},
				ActiveChallenges: []Challenge{
					{
						ChallengerID:      "5678",
						DefenderID:        "1234",
						ChallengeDate:     time.Now(),
						ChallengeDeadline: time.Now().Add(7 * 24 * time.Hour),
					},
				},
			},
		},
	}

	channel, err := data.findChannel("1234")
	if err != nil {
		t.Errorf("Error finding channel: %s", err)
		return
	}

	// Create guild member map (5678 left the server)
	guildMembers := map[string]bool{
		"1234": true,
		"9012": true,
		// 5678 is not in the map (left the server)
	}

	// Remove player 5678
	removed, err := channel.RemovePlayerIfNotInGuild("5678", guildMembers)
	if err != nil {
		t.Errorf("Error removing player: %s", err)
	}

	// Check that player was removed
	assert.Equal(t, removed, true)
	assert.Equal(t, len(channel.RankedPlayers), 2)

	// Check that the challenge involving player 5678 was removed
	assert.Equal(t, len(channel.ActiveChallenges), 0)

	// Check that positions were fixed
	for i := range channel.RankedPlayers {
		assert.Equal(t, channel.RankedPlayers[i].Position, i+1)
	}

	// Try to remove a player that is still in the guild
	removed, err = channel.RemovePlayerIfNotInGuild("1234", guildMembers)
	if err != nil {
		t.Errorf("Error removing player: %s", err)
	}
	assert.Equal(t, removed, false)
	assert.Equal(t, len(channel.RankedPlayers), 2)
}

func TestGetAllPlayerIDs(t *testing.T) {
	data := RankingData{
		Version: "v1_test",
		Channels: []*ChannelRankingData{
			{
				ChannelID: "1234",
				RankedPlayers: []Player{
					{PlayerID: "1234", GameName: "u1234", Status: "active", Position: 1},
					{PlayerID: "5678", GameName: "u5678", Status: "active", Position: 2},
					{PlayerID: "9012", GameName: "u9012", Status: "active", Position: 3},
				},
			},
		},
	}

	channel, err := data.findChannel("1234")
	if err != nil {
		t.Errorf("Error finding channel: %s", err)
		return
	}

	playerIDs := channel.GetAllPlayerIDs()
	assert.Equal(t, len(playerIDs), 3)

	// Check that all player IDs are present
	expectedIDs := map[string]bool{
		"1234": true,
		"5678": true,
		"9012": true,
	}
	for _, id := range playerIDs {
		if !expectedIDs[id] {
			t.Errorf("Unexpected player ID: %s", id)
		}
	}
}
