package kafka

import (
	kkafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

// Client Kafka Client
var Client ClientType = ClientType{}

// ClientType Kafka client type
type ClientType struct {
	Consumer *kkafka.Consumer
	Producer *kkafka.Producer
}

// Init Kafka
func Init(URL string, groupID string) error {
	consumer, err := kkafka.NewConsumer(&kkafka.ConfigMap{
		"bootstrap.servers": URL,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		return err
	}

	Client.Consumer = consumer

	producer, err := kkafka.NewProducer(&kkafka.ConfigMap{
		"bootstrap.servers": URL,
	})

	if err != nil {
		return err
	}

	Client.Producer = producer

	return nil
}

// Close Kafka client
func Close() {
	Client.Consumer.Close()
	Client.Producer.Close()
}
