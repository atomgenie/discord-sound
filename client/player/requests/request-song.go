package requests

import (
	"discord-sound/utils/kafka"
	"encoding/json"
	"os"

	kkafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

// RequestSong Request for a song
func RequestSong(query string, requestID string, server *Instance) error {

	topicURL := os.Getenv("YOUTUBE_DL_TOPIC")

	payload := kafka.YoutubeDLTopic{
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

	kafka.Client.Producer.Produce(&kkafka.Message{
		Value: payloadString,
		TopicPartition: kkafka.TopicPartition{
			Topic: &topicURL,
		},
	}, nil)

	return nil
}

// CancelRequest Cancel request
func CancelRequest(requestID string) {
	requestMux.Lock()
	requestMap[requestID] = request{}
	requestMux.Unlock()
}
