package requests

import (
	"discord-sound/utils/redis"
	"encoding/json"
	"os"
	"time"
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
	topicURLDone := os.Getenv("YOUTUBE_DL_DONE_TOPIC")
	messages, err := redis.SubscribePubSub(topicURLDone)

	if err != nil {
		panic(err)
	}

	for {
		msg := <-messages
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
