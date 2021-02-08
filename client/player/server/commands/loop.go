package commands

import (
	"discord-sound/player/server/guilds"

	"github.com/bwmarrin/discordgo"
)

// HandleLoop Handle loop command
func HandleLoop(s *discordgo.Session, g *guilds.Type, arg string) {
	var loopState guilds.LoopEnum = guilds.NoLoop

	if arg == "music" {
		loopState = guilds.LoopSound
	} else if arg == "queue" {
		loopState = guilds.LoopQueue
	} else {
		if g.GetLoop() == guilds.NoLoop {
			loopState = guilds.LoopQueue
		} else {
			loopState = guilds.NoLoop
		}
	}

	g.SetLoop(loopState)
}
