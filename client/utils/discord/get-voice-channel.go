package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// GetVoiceChannel Get the voice channel where the message author is connected
func GetVoiceChannel(s *discordgo.Session, m *discordgo.MessageCreate, guildID string) (string, error) {
	g, err := s.State.Guild(guildID)

	if err != nil {
		return "", nil
	}

	for _, voice := range g.VoiceStates {
		if voice.UserID == m.Author.ID {
			return voice.ChannelID, nil
		}
	}

	return "", fmt.Errorf("User not connected to any voice channel")
}
