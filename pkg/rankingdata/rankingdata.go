package rankingdata

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type RankingData struct {
	Version  string               `yaml:"version"`
	Channels []ChannelRankingData `yaml:"channels"`
}

type ChannelRankingData struct {
	ChannelID        string          `yaml:"channel_id"`
	RankedPlayers    []Player        `yaml:"ranked_players"`
	ActiveChallenges []Challenge     `yaml:"active_challenges"`
	ResultHistory    []ResultHistory `yaml:"result_history"`
}

type Player struct {
	Name              string `yaml:"name"`
	Status            string `yaml:"status"`
	Challenging       string `yaml:"challenging,optional"`
	Challenged        string `yaml:"challenged,optional"`
	ChallengeDate     string `yaml:"challenge_date,optional"`
	ChallengeDeadline string `yaml:"challenge_deadline,optional"`
	TimeZone          string `yaml:"time_zone,optional"`
	PrefferedServer   string `yaml:"preffered_server,optional"`
}

type Challenge struct {
	Challenger        string `yaml:"challenger"`
	Challenged        string `yaml:"challenged"`
	ChallengeDate     string `yaml:"challenge_date"`
	ChallengeDeadline string `yaml:"challenge_deadline"`
}

type ResultHistory struct {
	Challenger      string `yaml:"challenger"`
	Challenged      string `yaml:"challenged"`
	ChallengerScore int    `yaml:"challenger_score"`
	ChallengedScore int    `yaml:"challenged_score"`
	ChallengeDate   string `yaml:"challenge_date"`
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
func WriteRankingData(path string, rankingData *RankingData) error {
	yamlBytes, err := yaml.Marshal(*rankingData)
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

// function that prints a RankingData struct
func PrintRankingData(rankingData *RankingData) {
	yamlBytes, err := yaml.Marshal(*rankingData)
	if err != nil {
		panic(err)
	}
	yamlString := string(yamlBytes)
	println(yamlString)
}

// function that adds a new player to the ranking data
func AddPlayer(rankingData *RankingData, channelID string, playerName string) {
	for i, channelRankingData := range rankingData.Channels {
		if channelRankingData.ChannelID == channelID {
			rankingData.Channels[i].RankedPlayers = append(rankingData.Channels[i].RankedPlayers, Player{
				Name:   playerName,
				Status: "active",
			})
			return
		}
	}
	// if the channel is not found, add it to the ranking data
	rankingData.Channels = append(rankingData.Channels, ChannelRankingData{})
	rankingData.Channels[len(rankingData.Channels)-1].ChannelID = channelID
	rankingData.Channels[len(rankingData.Channels)-1].RankedPlayers = append(
		rankingData.Channels[len(rankingData.Channels)-1].RankedPlayers, Player{
			Name:   playerName,
			Status: "active",
		})
}

// function that removes a player from the ranking data
func RemovePlayer(rankingData *RankingData, channelID string, playerName string) {
	for i, channelRankingData := range rankingData.Channels {
		if channelRankingData.ChannelID == channelID {
			for j, player := range rankingData.Channels[i].RankedPlayers {
				if player.Name == playerName {
					rankingData.Channels[i].RankedPlayers = append(rankingData.Channels[i].RankedPlayers[:j], rankingData.Channels[i].RankedPlayers[j+1:]...)
					return
				}
			}
		}
	}
}
