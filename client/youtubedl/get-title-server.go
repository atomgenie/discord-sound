package youtubedl

import (
	"discord-sound/shared"
	"encoding/json"
	"net/http"
)

func handleGetTitle(res http.ResponseWriter, req *http.Request) {
	var data shared.GetTitlePayload

	err := json.NewDecoder(req.Body).Decode(&data)

	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := getID(data.Query)

	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	title, err := getTitle(id)

	if err != nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	response := shared.ResTitlePayload{}
	response.Title = title
	response.ID = id

	json.NewEncoder(res).Encode(response)
}

func startServer(addr string) {
	server := http.NewServeMux()
	server.HandleFunc("/youtube/getTitle", handleGetTitle)
	http.ListenAndServe(addr, server)
}
