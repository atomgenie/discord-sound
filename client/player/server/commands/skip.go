package commands

import (
	"discord-sound/player/server/guilds"
	"time"
)

// HandleSkip skip command
func HandleSkip(guild *guilds.Type) {

	guild.Mux.Lock()
	playing := guild.Playing
	guild.Mux.Unlock()

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
