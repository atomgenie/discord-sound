package kafka

// YoutubeDLTopic Topic
type YoutubeDLTopic struct {
	// ID id of message
	ID    string `json:"id"`
	Query string `json:"query"`
}

// YoutubeDLDoneTopic Topic
type YoutubeDLDoneTopic struct {
	// ID id of message
	ID         string `json:"id"`
	YoutubeID  string `json:"youtubeId"`
	MusicTitle string `json:"musicTitle"`
}
