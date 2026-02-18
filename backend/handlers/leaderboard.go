package handlers

import (
	"net/http"
	"pauls-bach/store"
	"sort"
)

type LeaderboardHandler struct {
	Store *store.Store
}

type leaderboardEntry struct {
	Rank     int    `json:"rank"`
	Username string `json:"username"`
	Balance  int    `json:"balance"`
	UserID   int    `json:"user_id"`
}

func (h *LeaderboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	store.ReadLock()
	defer store.ReadUnlock()

	users, err := h.Store.Users.GetAll()
	if err != nil {
		jsonError(w, "failed to load users", http.StatusInternalServerError)
		return
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].Balance > users[j].Balance
	})

	entries := make([]leaderboardEntry, 0, len(users))
	for i, u := range users {
		if u.IsAdmin {
			continue
		}
		entries = append(entries, leaderboardEntry{
			Rank:     i + 1,
			Username: u.Username,
			Balance:  u.Balance,
			UserID:   u.ID,
		})
	}
	// Re-rank after filtering admin
	for i := range entries {
		entries[i].Rank = i + 1
	}

	jsonResp(w, entries, http.StatusOK)
}
