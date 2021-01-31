package player

import (
	"discord-sound/utils/kafka"
	"encoding/json"
	"os"

	kkafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

// Start player
func Start() {
	kafkaURL := os.Getenv("KAFKA_URL")
	err := kafka.Init(kafkaURL, "player")

	if err != nil {
		panic(err)
	}

	defer kafka.Close()

	topicURL := os.Getenv("YOUTUBE_DL_TOPIC")

	payload := kafka.YoutubeDLTopic{
		ID:    "myId",
		Query: "Vald Gotaga",
	}

	payloadStr, _ := json.Marshal(payload)

	kafka.Client.Producer.Produce(&kkafka.Message{
		TopicPartition: kkafka.TopicPartition{Topic: &topicURL, Partition: kkafka.PartitionAny},
		Value:          payloadStr,
	}, nil)

	select {}

}
