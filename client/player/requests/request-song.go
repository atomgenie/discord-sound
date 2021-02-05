package requests

import (
	"context"
	"discord-sound/utils/redis"
	"encoding/json"
	"os"

	"github.com/mediocregopher/radix/v4"
)

// RequestSong Request for a song
func RequestSong(query string, requestID string, server *Instance) error {

	topicURL := os.Getenv("YOUTUBE_DL_TOPIC")

	payload := redis.YoutubeDLTopic{
		ID:    requestID,
		Query: query,
	}

	payloadString, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	requestMux.Lock()
	requestMap[requestID] = request{
		Server: server,
	}
	requestMux.Unlock()

	err = redis.Client.Client.Do(context.Background(), radix.Cmd(nil, "LPUSH", topicURL, string(payloadString)))

	return err
}

// CancelRequest Cancel request
func CancelRequest(requestID string) {
	requestMux.Lock()
	requestMap[requestID] = request{}
	requestMux.Unlock()
}
