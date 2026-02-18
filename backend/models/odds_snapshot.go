package models

type OddsSnapshot struct {
	ID        int     `json:"id"`
	EventID   int     `json:"event_id"`
	OutcomeID int     `json:"outcome_id"`
	Odds      float64 `json:"odds"`
	CreatedAt string  `json:"created_at"`
}
