package youtubedl

import (
	"discord-sound/utils/kafka"
	"discord-sound/utils/redis"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	kkafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

func downloadAndInstall() error {
	res, err := http.Get("https://yt-dl.org/downloads/latest/youtube-dl")

	if err != nil {
		return err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil
	}

	err = ioutil.WriteFile("youtube-dl", body, 0744)

	return err
}

func verifyInstall() error {
	checkExists := exec.Command("./youtube-dl", "--version")

	err := checkExists.Run()

	if err != nil {
		err := downloadAndInstall()
		return err
	}

	return nil
}

// Start Start youtubedl manager
func Start() {
	err := verifyInstall()

	if err != nil {
		panic(err)
	}

	redisURL := os.Getenv("REDIS_URL")
	err = redis.Init(redisURL)

	if err != nil {
		panic(err)
	}

	kafkaURL := os.Getenv("KAFKA_URL")
	err = kafka.Init(kafkaURL, "youtubedl")

	if err != nil {
		panic(err)
	}

	topicURL := os.Getenv("YOUTUBE_DL_TOPIC")
	topicURLDone := os.Getenv("YOUTUBE_DL_DONE_TOPIC")

	kafka.Client.Consumer.Subscribe(topicURL, nil)

	defer kafka.Close()

	fmt.Println("YoutubeSL Started")

	for {
		msg, err := kafka.Client.Consumer.ReadMessage(-1)
		if err != nil {
			fmt.Println("Consumer error", err)
		} else {

			var payload kafka.YoutubeDLTopic
			err := json.Unmarshal(msg.Value, &payload)

			if err != nil {
				continue
			}

			id, name, err := download(payload.Query)

			if err != nil {
				fmt.Println("Can't download music", err)
				continue
			}

			donePayload := kafka.YoutubeDLDoneTopic{
				ID:         payload.ID,
				MusicTitle: name,
				YoutubeID:  id,
			}

			donePayloadStr, err := json.Marshal(donePayload)

			if err != nil {
				continue
			}

			kafka.Client.Producer.Produce(&kkafka.Message{
				TopicPartition: kkafka.TopicPartition{
					Topic:     &topicURLDone,
					Partition: kkafka.PartitionAny,
				},
				Value: donePayloadStr,
			}, nil)

			runtime.GC()

		}
	}

}
