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

		if item.Title != "" {
			messageContent.WriteString("[")
			messageContent.WriteString(item.Title)
			messageContent.WriteString("](https://www.youtube.com/watch?v=")
			messageContent.WriteString(item.ID)
			messageContent.WriteString(")")
		} else {
			messageContent.WriteString(item.Query)
		}

		messageContent.WriteRune('\n')
	}

	nowPlaying := guild.GetNowPlaying()
	nowPlayingID := guild.GetNowPlayingID()

	nowPlayingText := ""

	if nowPlaying == "" || nowPlayingID == "" {
		nowPlayingText = "*Music is not being played*"
	} else {
		nowPlayingText = "[" + nowPlaying + "](https://www.youtube.com/watch?v=" + nowPlayingID + ")"
	}

	var status string

	if guild.GetPause() {
		status = "*Paused*"
	} else if guild.GetPlaying() {
		status = "*Playing*"
	} else {
		status = "*Inactive*"
	}

	var loopString string
	loopStatus := guild.GetLoop()

	switch loopStatus {
	case guilds.NoLoop:
		loopString = "*Not looping*"
	case guilds.LoopQueue:
		loopString = "*Looping queue*"
	case guilds.LoopSound:
		loopString = "*Looping music*"
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
					Value: nowPlayingText,
				},
				{
					Name:  "Queue",
					Value: messageContent.String(),
				},
				{
					Name:  "Loop",
					Value: loopString,
				},
			},
			Color: 37394,
			Title: "HxV",
		},
	})
}
