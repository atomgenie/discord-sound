package guilds

import (
	"sync"

	"github.com/bwmarrin/discordgo"
)

// QueueType queue type
type QueueType struct {
	Query string
	Title string
	ID    string
	UUID  string
}

// ResumePayload resume type
type ResumePayload struct {
	Message *discordgo.MessageCreate
}

// Type Guild type
type Type struct {
	ID             string
	playing        bool
	pause          bool
	queue          []QueueType
	Mux            sync.Mutex
	SoundChannelID string
	Skip           chan int
	PauseChan      chan int
	ResumeChan     chan ResumePayload
	StopChan       chan int
	nowPlaying     string
	voiceChannel   *discordgo.VoiceConnection
}

// Map Guilds maps
var Map map[string]*Type = make(map[string]*Type)

// Mux Guilds mutex
var Mux sync.Mutex

// New instance
func New(guildID string) *Type {
	guildInstance := new(Type)
	guildInstance.ID = guildID
	guildInstance.playing = false
	guildInstance.pause = false
	guildInstance.queue = make([]QueueType, 0)
	guildInstance.Skip = make(chan int)
	guildInstance.PauseChan = make(chan int)
	guildInstance.ResumeChan = make(chan ResumePayload)
	guildInstance.StopChan = make(chan int)

	return guildInstance
}

// GetPlaying get playing
func (g *Type) GetPlaying() bool {
	g.Mux.Lock()
	defer g.Mux.Unlock()
	return g.playing

}

// SetPlaying set playing
func (g *Type) SetPlaying(value bool) {
	g.Mux.Lock()
	defer g.Mux.Unlock()

	g.playing = value
}

// GetPause Get pause status
func (g *Type) GetPause() bool {
	g.Mux.Lock()
	defer g.Mux.Unlock()
	return g.pause
}

// GetQueue Get queue
func (g *Type) GetQueue() []QueueType {
	g.Mux.Lock()
	defer g.Mux.Unlock()
	queue := g.queue
	return queue
}

// GetNowPlaying Get now playing
func (g *Type) GetNowPlaying() string {
	g.Mux.Lock()
	nowPlaying := g.nowPlaying
	g.Mux.Unlock()

	return nowPlaying
}

// SetPause Set pause
func (g *Type) SetPause(pause bool) {
	g.Mux.Lock()
	defer g.Mux.Unlock()
	g.pause = pause
}

// QueueAppend Append to queue
func (g *Type) QueueAppend(element QueueType) {
	g.Mux.Lock()
	defer g.Mux.Unlock()
	g.queue = append(g.queue, element)
}

// QueuePopFront pop front element from queue and return it.
// If empty, bool is set to true
// frontElement, isEmpty
func (g *Type) QueuePopFront() (QueueType, bool) {
	g.Mux.Lock()
	defer g.Mux.Unlock()

	if len(g.queue) == 0 {
		return QueueType{}, true
	}

	firstElement := g.queue[0]
	g.queue = g.queue[1:]

	return firstElement, false
}

// SetQueueTitle Set title of a element in queue
func (g *Type) SetQueueTitle(query string, title string, ID string) {
	g.Mux.Lock()
	defer g.Mux.Unlock()

	for i, elm := range g.queue {
		if elm.Query == query && elm.Title == "" {
			g.queue[i].Title = title
			g.queue[i].ID = ID
			break
		}
	}
}

// QueueLen get queue length
func (g *Type) QueueLen() int {
	g.Mux.Lock()
	defer g.Mux.Unlock()

	return len(g.queue)
}

// SetNowPlaying set nowPlaying
func (g *Type) SetNowPlaying(nowPLaying string) {
	g.Mux.Lock()
	defer g.Mux.Unlock()

	g.nowPlaying = nowPLaying
}

// SetVoiceChannel Set voice channel
func (g *Type) SetVoiceChannel(value *discordgo.VoiceConnection) {
	g.Mux.Lock()
	defer g.Mux.Unlock()
	g.voiceChannel = value
}

// GetVoiceChannel Get voice channel
func (g *Type) GetVoiceChannel() *discordgo.VoiceConnection {
	g.Mux.Lock()
	defer g.Mux.Unlock()

	return g.voiceChannel
}
