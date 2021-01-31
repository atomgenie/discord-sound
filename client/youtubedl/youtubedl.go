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

	kafka.Client.Consumer.Subscribe(topicURL, nil)

	defer kafka.Close()

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

			_, err = download(payload.Query)

			if err != nil {
				fmt.Println("Can't download music", err)
				continue
			}

		}
	}

}
