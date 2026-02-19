package models

type ActivityEntry struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	UserID    int    `json:"user_id"`
	EventID   int    `json:"event_id"`
	CreatedAt string `json:"created_at"`
}
