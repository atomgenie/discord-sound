package commands

import (
	"discord-sound/player/server/guilds"
	"time"

	"github.com/bwmarrin/discordgo"
)

// HandleResume command
func HandleResume(g *guilds.Type, m *discordgo.MessageCreate) {
	if !g.GetPause() {
		return
	}

	select {
	case g.ResumeChan <- guilds.ResumePayload{
		Message: m,
	}:
		break
	case <-time.After(5 * time.Second):
		break
	}
}
