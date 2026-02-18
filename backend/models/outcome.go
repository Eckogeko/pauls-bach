package models

type Outcome struct {
	ID      int    `json:"id"`
	EventID int    `json:"event_id"`
	Label   string `json:"label"`
}
