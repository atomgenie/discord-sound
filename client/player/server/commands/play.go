package commands

import (
	"discord-sound/player/server/guilds"
	"discord-sound/utils/discord"

	"github.com/bwmarrin/discordgo"
)

// HandlePlay Handle play command
func HandlePlay(s *discordgo.Session, m *discordgo.MessageCreate, argument string, guild *guilds.Type) {

	if argument == "" {
		return
	}

	playing := guild.GetPlaying()

	if !playing {

		voiceID, err := discord.GetVoiceChannel(s, m, guild.ID)

		if err != nil {
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
