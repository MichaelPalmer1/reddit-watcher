package redditbot

import (
	"fmt"
	"strings"

	"github.com/MichaelPalmer1/reddit-watcher/config"
	"github.com/bwmarrin/discordgo"
	"github.com/turnage/graw"
	"github.com/turnage/graw/reddit"
)

// screenzBot struct
type screenzBot struct {
	Bot            reddit.Bot
	DiscordSession *discordgo.Session
}

// Post to Discord when something is posted to Reddit
func (r *screenzBot) Post(post *reddit.Post) error {
	// List all guilds the bot is a member of
	for _, guild := range r.DiscordSession.State.Guilds {
		// List all channels in this guild
		channels, err := r.DiscordSession.GuildChannels(guild.ID)
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
				} else if post.Domain == "youtube.com" || post.Domain == "youtu.be" {
					video := discordgo.MessageEmbedVideo{
						URL: post.URL,
					}
					embed.Video = &video
					embed.URL = "https://reddit.com" + post.Permalink
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
				r.DiscordSession.ChannelMessageSendComplex(channel.ID, &msg)

				// r.dg.ChannelMessageSend(channel.ID, fmt.Sprintf("*%s*\n\n%s\n\n```\n%s\n```", post.Title, post.Author, post.SelfText))
			}
		}
	}

	fmt.Printf("%s posted \"%s\" in r/%s\n", post.Author, post.Title, post.Subreddit)
	return nil
}

// StartReddit - Starts up the reddit event watcher
func StartReddit(session *discordgo.Session, conf *config.Config) {
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
	handler := &screenzBot{Bot: bot, DiscordSession: session}
	if _, wait, err := graw.Run(handler, bot, cfg); err != nil {
		fmt.Println("Failed to start graw run: ", err)
	} else {
		fmt.Println("graw run failed: ", wait())
	}
}
