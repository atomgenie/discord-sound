package requests

import (
	"context"
	"discord-sound/utils/redis"
	"encoding/json"
	"os"
	"time"

	"github.com/mediocregopher/radix/v4"
)

// DoneStruct done
type DoneStruct struct {
	YoutubeID    string
	YoutubeTitle string
}

// Instance Instance of a server
type Instance struct {
	DoneChan chan DoneStruct
}

func handleDone(ID string, youtubeID string, youtubeTitle string) {

	requestMux.Lock()
	targetRequest := requestMap[ID]
	requestMux.Unlock()

	if targetRequest.Server == nil {
		return
	}

	select {

	case targetRequest.Server.DoneChan <- DoneStruct{
		YoutubeID:    youtubeID,
		YoutubeTitle: youtubeTitle,
	}:
		break
	case <-time.After(10 * time.Second):
		break
	}

	requestMap[ID] = request{}
}

func handleDoneRequests() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	topicURLDone := os.Getenv("YOUTUBE_DL_DONE_TOPIC")
	redisURL := os.Getenv("REDIS_URL")

	conn, err := radix.Dial(context.Background(), "tcp", redisURL)

	if err != nil {
		panic(err)
	}

	subRedis := (radix.PubSubConfig{}).New(conn)
	chanSub := make(chan radix.PubSubMessage)

	if err := subRedis.Subscribe(ctx, chanSub, topicURLDone); err != nil {
		panic(err)
	}

	for {
		msg := <-chanSub

		var payload redis.YoutubeDLDoneTopic
		err := json.Unmarshal(msg.Message, &payload)

		if err != nil {
			continue
		}

		handleDone(payload.ID, payload.YoutubeID, payload.MusicTitle)
	}
}

// StartDone Start done mecanism
func StartDone() {
	go handleDoneRequests()
}
