package commands

import (
	"discord-sound/player/server/guilds"
	"time"
)

// HandleSkip skip command
func HandleSkip(guild *guilds.Type) {
	playing := guild.GetPlaying()

	if !playing {
		return
	}

	select {
	case guild.Skip <- 1:
		break
	case <-time.After(5 * time.Second):
		break

	}
}
