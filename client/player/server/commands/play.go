package commands

import (
	"discord-sound/player/server/guilds"

	"github.com/bwmarrin/discordgo"
)

// HandlePlay Handle play command
func HandlePlay(s *discordgo.Session, m *discordgo.MessageCreate, argument string, guild *guilds.Type) {

	if argument == "" {
		return
	}

	playing := guild.GetPlaying()

	if !playing {
		g, err := s.State.Guild(guild.ID)

		if err != nil {
			return
		}

		var voiceID string = ""

		for _, voice := range g.VoiceStates {
			if voice.UserID == m.Message.Author.ID {
				voiceID = voice.ChannelID
				break
			}
		}

		if voiceID == "" {
			return
		}

		guild.Mux.Lock()
		guild.SoundChannelID = voiceID
		guild.Mux.Unlock()
	}

	addToQueue(argument, guild)

	if !playing {
		go startQueue(s, guild)
	}
}
