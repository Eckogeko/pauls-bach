package models

type Event struct {
	ID               int    `json:"id"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	EventType        string `json:"event_type"` // "binary" or "multi"
	Status           string `json:"status"`      // "open", "closed", "resolved"
	WinningOutcomeID int    `json:"winning_outcome_id,omitempty"`
	CreatedAt        string `json:"created_at"`
	ResolvedAt       string `json:"resolved_at,omitempty"`
	CreatorID        int    `json:"creator_id,omitempty"`
	BountyPaid       bool   `json:"bounty_paid,omitempty"`
}
