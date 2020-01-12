package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/turnage/graw"
	"github.com/turnage/graw/reddit"
)

type screenzBot struct {
	bot reddit.Bot
	dg  *discordgo.Session
}

type config struct {
	Token           string
	RedditAppID     string
	RedditAppSecret string
	Subreddits      string
}

/**
 * Post to Discord when something is posted to Reddit
 */
func (r *screenzBot) Post(post *reddit.Post) error {
	// List all guilds the bot is a member of
	for _, guild := range r.dg.State.Guilds {
		// List all channels in this guild
		channels, err := r.dg.GuildChannels(guild.ID)
		if err != nil {
			fmt.Println("guild channel error", err)
			return err
		}

		// Find the reddit category
		var redditCategory string
		for _, channel := range channels {
			// Find reddit category
			if channel.Type == discordgo.ChannelTypeGuildCategory && channel.Name == "reddit" {
				redditCategory = channel.ID
				break
			}
		}

		for _, channel := range channels {
			// Skip all non-text channels
			if channel.Type != discordgo.ChannelTypeGuildText {
				continue
			}

			// Skip all channels that are not children of the reddit category
			if channel.ParentID != redditCategory {
				continue
			}

			// Only post if the channel name matches the subreddit name
			if channel.Name == post.Subreddit {
				// Build the author
				author := discordgo.MessageEmbedAuthor{
					Name: post.Author,
				}

				// Build the footer
				footer := discordgo.MessageEmbedFooter{
					Text: post.LinkFlairText,
				}

				// Build the embed
				embed := discordgo.MessageEmbed{
					Title:       post.Title,
					Author:      &author,
					Footer:      &footer,
					Description: post.SelfText,
				}

				// Check if this is a self post
				if post.IsSelf {
					embed.URL = post.URL
				} else {
					// Build the image
					image := discordgo.MessageEmbedImage{
						URL: post.URL,
					}
					embed.Image = &image
					embed.URL = "https://reddit.com" + post.Permalink
				}

				// Build the message
				msg := discordgo.MessageSend{
					Embed: &embed,
				}

				// Send the message
				r.dg.ChannelMessageSendComplex(channel.ID, &msg)

				// r.dg.ChannelMessageSend(channel.ID, fmt.Sprintf("*%s*\n\n%s\n\n```\n%s\n```", post.Title, post.Author, post.SelfText))
			}
		}
	}

	fmt.Printf("%s posted \"%s\" in %s\n", post.Author, post.Title, post.Subreddit)
	return nil
}

/**
 * Start up the reddit event watcher
 */
func startReddit(dg *discordgo.Session, conf *config) {
	// Build the bot config
	botConfig := reddit.BotConfig{
		Agent: "graw:test_bot:0.3.1",
		App: reddit.App{
			ID:     conf.RedditAppID,
			Secret: conf.RedditAppSecret,
		},
	}

	// Create the bot
	bot, err := reddit.NewBot(botConfig)
	if err != nil {
		fmt.Println("error talking to reddit,", err)
		return
	}

	// Configure the subreddits to track
	subreddits := strings.Split(conf.Subreddits, ",")
	cfg := graw.Config{Subreddits: subreddits}

	fmt.Println("watching subreddits")

	// Initialize the bot
	handler := &screenzBot{bot: bot, dg: dg}
	if _, wait, err := graw.Run(handler, bot, cfg); err != nil {
		fmt.Println("Failed to start graw run: ", err)
	} else {
		fmt.Println("graw run failed: ", wait())
	}
}

/**
 * ACK to all messages sent by users
 */
func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == session.State.User.ID {
		return
	}

	fmt.Println("Message received:", message)

	session.ChannelMessageSend(message.ChannelID, "ACK")
}

func main() {
	// Command line arguments
	var conf config
	flag.StringVar(&conf.Token, "token", "", "Bot Token")
	flag.StringVar(&conf.RedditAppID, "app-id", "", "Reddit app id")
	flag.StringVar(&conf.RedditAppSecret, "app-secret", "", "Reddit app secret")
	flag.StringVar(&conf.Subreddits, "subreddits", "funny,homelab,homelabsales,hardwareswap,memes", "Subreddits to watch")
	flag.Parse()

	// Create discord session
	dg, err := discordgo.New("Bot " + conf.Token)
	if err != nil {
		fmt.Println("error encountered while creating session,", err)
		return
	}

	// Register handler
	dg.AddHandler(messageCreate)
	dg.AddHandler(func(discord *discordgo.Session, ready *discordgo.Ready) {
		discord.UpdateStatus(0, "")
		guilds := discord.State.Guilds
		fmt.Println("Ready with", len(guilds), "guilds.")
	})

	// Open websocket
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening session,", err)
		return
	}

	// Block until TERM signal received
	fmt.Println("Started bot")
	startReddit(dg, &conf)
	<-make(chan struct{})
}
