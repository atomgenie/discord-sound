package discord

import "github.com/bwmarrin/discordgo"

var readyChan chan int = make(chan int)

// ClientType Client type
type ClientType struct {
	Client *discordgo.Session
}

// Client Discord client
var Client ClientType

// Init Discord API
func Init(token string) error {

	dg, err := discordgo.New("Bot " + token)

	if err != nil {
		return err
	}

	dg.AddHandler(ready)

	Client.Client = dg

	return nil
}

// Open client
func Open() error {
	err := Client.Client.Open()

	if err != nil {
		return err
	}

	<-readyChan

	return nil
}

// Close client
func Close() {
	Client.Client.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "Ready!")

	readyChan <- 1

}
