package commands

import (
	"discord-sound/player/server/guilds"
	"time"
)

// HandleResume command
func HandleResume(g *guilds.Type) {
	if !g.GetPlaying() || !g.GetPause() {
		return
	}

	select {
	case g.ResumeChan <- 1:
		break
	case <-time.After(5 * time.Second):
		break
	}
}
