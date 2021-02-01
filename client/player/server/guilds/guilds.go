package guilds

import (
	"sync"
)

// Type Guild type
type Type struct {
	ID             string
	Playing        bool
	Queue          []string
	Mux            sync.Mutex
	SoundChannelID string
}

// Map Guilds maps
var Map map[string]*Type = make(map[string]*Type)

// Mux Guilds mutex
var Mux sync.Mutex
