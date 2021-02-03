package commands

import (
	"discord-sound/player/server/guilds"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// HandleQueue Queue command
func HandleQueue(guild *guilds.Type, channelID string, s *discordgo.Session) {
	queue := guild.GetQueue()

	var messageContent strings.Builder

	if len(queue) == 0 {
		messageContent.WriteString("*Empty queue*")
	}

	for _, item := range queue {
		messageContent.WriteString("- ")
		messageContent.WriteString(item.Query)
		messageContent.WriteRune('\n')
	}

	nowPlaying := guild.GetNowPlaying()

	if nowPlaying == "" {
		nowPlaying = "*Music is not being played*"
	}

	var status string

	if guild.GetPause() {
		status = "*Paused*"
	} else if guild.GetPlaying() {
		status = "*Playing*"
	} else {
		status = "*Inactive*"
	}

	s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:  "Status",
					Value: status,
				},
				{
					Name:  "Now Playing",
					Value: nowPlaying,
				},
				{
					Name:  "Queue",
					Value: messageContent.String(),
				},
			},
			Color: 37394,
			Title: "Queue",
		},
	})
}
