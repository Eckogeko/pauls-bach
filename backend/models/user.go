package models

type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	PinHash   string `json:"-"`
	Balance   int    `json:"balance"`
	IsAdmin   bool   `json:"is_admin"`
	Bingo     bool   `json:"bingo"`
	CreatedAt string `json:"created_at"`
}
