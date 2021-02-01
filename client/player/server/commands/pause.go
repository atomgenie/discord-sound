package commands

import (
	"discord-sound/player/server/guilds"
	"time"
)

//HandlePause Handle pause
func HandlePause(g *guilds.Type) {
	if !g.GetPlaying() {
		return
	}

	select {
	case g.PauseChan <- 1:
		break
	case <-time.After(5 * time.Second):
		break
	}
}
