package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"discord_ladder_bot/pkg/config"
	"discord_ladder_bot/pkg/discordbot"
)

func main() {
	configPathPtr := flag.String("config", "config.yml", "Path to the config file.")
	rankingPathPtr := flag.String("ranking", "ranking.yml", "Path to the ranking file.")
	flag.Parse()

	conf, err := config.ReadConfig(*configPathPtr)
	if err != nil {
		panic(err)
	}

	discord, err := discordbot.NewDiscordBot(conf.Token, *rankingPathPtr)
	if err != nil {
		panic(err)
	}
	discord.Start()
	defer discord.Stop()

	fmt.Println("token: " + conf.Token)
	fmt.Println("ladder mode: " + conf.LadderMode)
	fmt.Println("ranking file path: " + *rankingPathPtr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press Ctrl+C to exit")
	<-stop
	signal.Reset(os.Interrupt)
	fmt.Println("Exiting...")
}
