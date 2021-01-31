package player

import (
	"discord-sound/utils/kafka"
	"discord-sound/utils/uuid"
	"encoding/json"
	"fmt"
	"os"

	kkafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

// Start player
func Start() {
	kafkaURL := os.Getenv("KAFKA_URL")
	err := kafka.Init(kafkaURL, "player-"+uuid.Gen())

	if err != nil {
		panic(err)
	}

	defer kafka.Close()

	topicURL := os.Getenv("YOUTUBE_DL_TOPIC")

	payload := kafka.YoutubeDLTopic{
		ID:    uuid.Gen(),
		Query: "Vald Gotaga",
	}

	payloadStr, _ := json.Marshal(payload)

	kafka.Client.Producer.Produce(&kkafka.Message{
		TopicPartition: kkafka.TopicPartition{Topic: &topicURL, Partition: kkafka.PartitionAny},
		Value:          payloadStr,
	}, nil)

	fmt.Println("Player Started")

	select {}

}
