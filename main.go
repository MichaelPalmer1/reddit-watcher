package main

import (
	"flag"
	"fmt"

	"github.com/MichaelPalmer1/reddit-watcher/config"
	"github.com/MichaelPalmer1/reddit-watcher/discordbot"
	"github.com/MichaelPalmer1/reddit-watcher/redditbot"
)

func main() {
	// Command line arguments
	var conf config.Config
	flag.StringVar(&conf.Token, "token", "", "Bot token")
	flag.StringVar(&conf.RedditAppID, "app-id", "", "Reddit app id")
	flag.StringVar(&conf.RedditAppSecret, "app-secret", "", "Reddit app secret")
	flag.StringVar(&conf.Subreddits, "subreddits", "funny,homelab,homelabsales,hardwareswap,memes", "Subreddits to watch")
	flag.Parse()

	// Start the discord bot
	session := discordbot.StartDiscord(&conf)
	if session == nil {
		fmt.Println("Error starting discord bot")
		return
	}

	// Start the reddit watcher
	redditbot.StartReddit(session, &conf)

	// Open a channel
	<-make(chan struct{})
}
