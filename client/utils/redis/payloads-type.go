package redis

// YoutubeDLDoneTopic Done type
type YoutubeDLDoneTopic struct {
	ID         string `json:"id"`
	MusicTitle string `json:"music_title"`
	YoutubeID  string `json:"youtubeID"`
}

// YoutubeDLTopic Type
type YoutubeDLTopic struct {
	ID    string `json:"id"`
	Query string `json:"query"`
}
