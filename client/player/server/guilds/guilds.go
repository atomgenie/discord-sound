package guilds

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

// QueueType queue type
type QueueType struct {
	Query string
	UUID  string
}

type ResumePayload struct {
	Message *discordgo.MessageCreate
}

// Type Guild type
type Type struct {
	ID             string
	Playing        bool
	Pause          bool
	Queue          []QueueType
	Mux            sync.Mutex
	SoundChannelID string
	Skip           chan int
	PauseChan      chan int
	ResumeChan     chan ResumePayload
	StopChan       chan int
	NowPlaying     string
}

// Map Guilds maps
var Map map[string]*Type = make(map[string]*Type)

// Mux Guilds mutex
var Mux sync.Mutex

// New instance
func New(guildID string) *Type {
	guildInstance := new(Type)
	guildInstance.ID = guildID
	guildInstance.Playing = false
	guildInstance.Pause = false
	guildInstance.Queue = make([]QueueType, 0)
	guildInstance.Skip = make(chan int)
	guildInstance.PauseChan = make(chan int)
	guildInstance.ResumeChan = make(chan ResumePayload)
	guildInstance.StopChan = make(chan int)

	return guildInstance
}

// GetPlaying get playing
func (g *Type) GetPlaying() bool {
	g.Mux.Lock()
	loading := g.Playing
	g.Mux.Unlock()

	return loading
}

// GetPause Get pause status
func (g *Type) GetPause() bool {
	g.Mux.Lock()
	pause := g.Pause
	g.Mux.Unlock()

	return pause
}

// GetQueue Get queue
func (g *Type) GetQueue() []QueueType {
	g.Mux.Lock()
	queue := g.Queue
	g.Mux.Unlock()
	return queue
}

// GetNowPlaying Get now playing
func (g *Type) GetNowPlaying() string {
	g.Mux.Lock()
	nowPlaying := g.NowPlaying
	g.Mux.Unlock()

	return nowPlaying
}
