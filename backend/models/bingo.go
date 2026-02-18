package models

type BingoEvent struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Resolved  bool   `json:"resolved"`
	CreatedAt string `json:"created_at"`
}

type BingoSquare struct {
	Position     int    `json:"position"`
	BingoEventID int    `json:"bingo_event_id,omitempty"`
	CustomText   string `json:"custom_text,omitempty"`
	Resolved     bool   `json:"resolved"`
}

type BingoBoard struct {
	ID        int           `json:"id"`
	UserID    int           `json:"user_id"`
	Squares   []BingoSquare `json:"squares"`
	CreatedAt string        `json:"created_at"`
}

type BingoWinner struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	BoardID   int    `json:"board_id"`
	Line      string `json:"line"`
	CreatedAt string `json:"created_at"`
}
