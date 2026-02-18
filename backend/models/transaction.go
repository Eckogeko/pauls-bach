package models

type Transaction struct {
	ID        int     `json:"id"`
	UserID    int     `json:"user_id"`
	EventID   int     `json:"event_id"`
	OutcomeID int     `json:"outcome_id"`
	TxType    string  `json:"tx_type"` // "buy", "sell", "payout", "bonus"
	Shares    float64 `json:"shares"`
	Points    int     `json:"points"`
	CreatedAt string  `json:"created_at"`
}
