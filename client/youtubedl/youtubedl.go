package youtubedl

import (
	"discord-sound/utils/redis"
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

	download("Vald Gotaga")

	select {}
}
