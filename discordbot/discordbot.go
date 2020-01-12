package discordbot

import (
	"fmt"

	"github.com/MichaelPalmer1/reddit-watcher/config"
	"github.com/bwmarrin/discordgo"
)

// StartDiscord - starts the discord bot
func StartDiscord(conf *config.Config) *discordgo.Session {
	// Create discord session
	session, err := discordgo.New("Bot " + conf.Token)
	if err != nil {
		fmt.Println("error encountered while creating session,", err)
		return nil
	}

	// Register handler
	session.AddHandler(messageCreate)
	session.AddHandler(func(discord *discordgo.Session, ready *discordgo.Ready) {
		discord.UserUpdateStatus(discordgo.StatusOnline)
		guilds := discord.State.Guilds
		fmt.Println("Ready with", len(guilds), "guilds.")
	})

	// Open websocket
	err = session.Open()
	if err != nil {
		fmt.Println("error opening session,", err)
		return nil
	}
	fmt.Println("Started bot")

	return session
}

// messageCreate - ACK to all messages sent by users
func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == session.State.User.ID {
		return
	}

	fmt.Println("Message received:", message)

	session.ChannelMessageSend(message.ChannelID, "ACK")
}
