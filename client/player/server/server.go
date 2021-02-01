package server

import (
	"discord-sound/player/server/commands"
	"discord-sound/player/server/guilds"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const command = "!!"

// HandleMessage Handler for message
func HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	channel, err := s.State.Channel(m.ChannelID)

	if err != nil {
		return
	}

	guild, err := s.State.Guild(channel.GuildID)

	if err != nil {
		return
	}

	guilds.Mux.Lock()

	actualGuild := guilds.Map[guild.ID]

	if actualGuild == nil {
		guildInstance := new(guilds.Type)
		guildInstance.ID = guild.ID
		guildInstance.Playing = false
		guildInstance.Queue = make([]string, 0)
		guilds.Map[guild.ID] = guildInstance
		actualGuild = guildInstance
	}

	guilds.Mux.Unlock()

	if strings.HasPrefix(m.Content, command) {
		firstArgs := m.Content[len(command):]

		fmt.Println("Command", firstArgs)

		if strings.HasPrefix(firstArgs, "play") {
			args := firstArgs[5:]

			commands.HandlePlay(s, m, args, actualGuild)
		}

	}
}
