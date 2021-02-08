package requests

import (
	"bytes"
	"discord-sound/shared"
	"encoding/json"
	"net/http"
	"os"
)

// RequestTitleRes Response type
type RequestTitleRes struct {
	ID    string
	Title string
}

// RequestTitle Request title to youtube-dl service
func RequestTitle(query string) (RequestTitleRes, error) {
	apiURL := os.Getenv("YOUTUBE_URL")
	toReturn := RequestTitleRes{}
	dataPayload := shared.GetTitlePayload{
		Query: query,
	}

	rawData, err := json.Marshal(dataPayload)

	if err != nil {
		return toReturn, err
	}

	dataBuffer := bytes.NewBuffer(rawData)
	resp, err := http.Post(apiURL+"/youtube/getTitle", "application/json", dataBuffer)

	if err != nil {
		return toReturn, err
	}

	defer resp.Body.Close()

	var respJSON shared.ResTitlePayload
	err = json.NewDecoder(resp.Body).Decode(&respJSON)

	if err != nil {
		return toReturn, err
	}

	toReturn.ID = respJSON.ID
	toReturn.Title = respJSON.Title

	return toReturn, nil
}
