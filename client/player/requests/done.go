package requests

import (
	"discord-sound/utils/kafka"
	"encoding/json"
	"fmt"
	"os"
)

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

	targetRequest.Server.DoneChan <- DoneStruct{
		YoutubeID:    youtubeID,
		YoutubeTitle: youtubeTitle,
	}

	requestMap[ID] = request{}
}

func handleDoneRequests() {

	topicURLDone := os.Getenv("YOUTUBE_DL_DONE_TOPIC")
	kafka.Client.Consumer.Subscribe(topicURLDone, nil)
	defer kafka.Close()

	for {
		msg, err := kafka.Client.Consumer.ReadMessage(-1)

		if err != nil {
			fmt.Println("Consumer error", err)
			continue
		}

		var payload kafka.YoutubeDLDoneTopic
		err = json.Unmarshal(msg.Value, &payload)

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
