package models

type Position struct {
	ID        int     `json:"id"`
	UserID    int     `json:"user_id"`
	EventID   int     `json:"event_id"`
	OutcomeID int     `json:"outcome_id"`
	Shares    float64 `json:"shares"`
	AvgPrice  float64 `json:"avg_price"`
	CreatedAt string  `json:"created_at"`
}
