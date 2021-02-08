package shared

// GetTitlePayload Payload
type GetTitlePayload struct {
	Query string `json:"query"`
}

// ResTitlePayload Payload
type ResTitlePayload struct {
	Title string `json:"title"`
	ID    string `json:"id"`
}
