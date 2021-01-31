package player

import (
	"discord-sound/utils/kafka"
	"os"

	kkafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

// Start player
func Start() {
	kafkaURL := os.Getenv("KAFKA_URL")
	err := kafka.Init(kafkaURL)

	if err != nil {
		panic(err)
	}

	defer kafka.Close()

	topicURL := os.Getenv("YOUTUBE_DL_TOPIC")

	kafka.Client.Producer.Produce(&kkafka.Message{
		TopicPartition: kkafka.TopicPartition{Topic: &topicURL, Partition: kkafka.PartitionAny},
		Value:          []byte("Gotaga vald"),
	}, nil)

	select {}

}
