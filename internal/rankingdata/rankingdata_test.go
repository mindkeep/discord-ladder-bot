package rankingdata

import (
	"testing"

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

	if err := channel.AddPlayer("1111", "u1111"); err != nil {
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

	if err := channel.RemovePlayer("5678"); err != nil {
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
	if err := channel.RemovePlayer("1111"); err == nil {
		t.Errorf("Error removing player: %s", err)
	}
	assert.Equal(t, len(channel.RankedPlayers), 4)
}
