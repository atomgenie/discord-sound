package youtubedl

import (
	"context"
	"discord-sound/utils/redis"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/mediocregopher/radix/v4"
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

	topicURL := os.Getenv("YOUTUBE_DL_TOPIC")
	topicURLDone := os.Getenv("YOUTUBE_DL_DONE_TOPIC")

	// sub, pubChan := radix.NewPubSubStubConn("tcp", redisURL, func(c context.Context, s []string) interface{} {
	// 	fmt.Println(s)
	// 	return nil
	// })

	// defer sub.Close()

	fmt.Println("YoutubeSL Started")

	var rawMessage []string

	for {
		err = redis.Client.Client.Do(context.Background(), radix.Cmd(&rawMessage, "BRPOP", topicURL, "0"))
		if err != nil {
			fmt.Println("Consumer error", err)
		} else {

			var payload redis.YoutubeDLTopic
			err = json.Unmarshal([]byte(rawMessage[1]), &payload)

			if err != nil {
				continue
			}

			id, name, err := download(payload.Query)

			if err != nil {
				fmt.Println("Can't download music", err)
				continue
			}

			donePayload := redis.YoutubeDLDoneTopic{
				ID:         payload.ID,
				MusicTitle: name,
				YoutubeID:  id,
			}

			donePayloadStr, err := json.Marshal(donePayload)

			if err != nil {
				continue
			}

			redis.Client.Client.Do(context.Background(), radix.Cmd(nil, "PUBLISH", topicURLDone, string(donePayloadStr)))
			runtime.GC()
		}
	}

}
