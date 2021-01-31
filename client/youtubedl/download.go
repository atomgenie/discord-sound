package youtubedl

import (
	"bytes"
	"context"
	"discord-sound/utils/redis"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/mediocregopher/radix/v4"
)

func getID(query string) (string, error) {
	r := exec.Command("./youtube-dl", "--get-id", "ytsearch:"+query)
	var out bytes.Buffer
	r.Stdout = &out
	err := r.Run()

	if err != nil {
		return "", err
	}

	data := out.String()
	return strings.TrimSuffix(data, "\n"), nil
}

func handleEndDownload(id string, filename string) error {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return err
	}

	err = redis.Client.Client.Do(context.Background(), radix.Cmd(nil, "SETEX", id, "1800", string(data)))

	if err != nil {
		return err
	}
	return nil
}

func download(query string) (string, error) {

	id, err := getID(query)

	fmt.Println("Downloading", query, id)

	if err != nil {
		return "", err
	}

	var exists bool
	err = redis.Client.Client.Do(context.Background(), radix.Cmd(&exists, "EXISTS", id))

	if err != nil {
		panic(err)
	}

	if exists {
		return id, nil
	}

	filename := id + ".webm"

	r := exec.Command("./youtube-dl", "--format", "bestaudio[ext=webm]", "-o", filename, id)
	err = r.Run()

	if err != nil {
		return "", err
	}

	defer os.Remove(filename)
	err = handleEndDownload(id, filename)

	return id, nil

}
