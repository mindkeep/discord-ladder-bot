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

	// Read the initial config file.
	// Note: When run from a container, this will need to be mounted in as a secret.
	configPathPtr := flag.String("config", "config.yml", "Path to the config file.")

	// TODO: convert this into a function that takes the mongodb key and returns a pointer to the database.
	rankingPathPtr := flag.String("ranking", "ranking.yml", "Path to the ranking file.")

	flag.Parse()

	conf, err := config.ReadConfig(*configPathPtr)
	if err != nil {
		panic(err)
	}

	// TODO: pass in database pointer and maybe OpenAI client pointer.
	discord, err := discordbot.NewDiscordBot(conf.DiscordToken, *rankingPathPtr)
	if err != nil {
		panic(err)
	}
	discord.Start()
	defer discord.Stop()

	// Gracefully Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	fmt.Println("Press Ctrl+C to exit")
	<-stop
	signal.Reset(os.Interrupt)
	fmt.Println("Exiting...")
}
