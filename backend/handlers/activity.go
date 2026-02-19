package handlers

import (
	"net/http"
	"pauls-bach/models"
	"pauls-bach/store"
)

type ActivityHandler struct {
	Store *store.Store
}

func (h *ActivityHandler) GetRecent(w http.ResponseWriter, r *http.Request) {
	store.ReadLock()
	defer store.ReadUnlock()

	all, err := h.Store.Activity.GetRecent(100)
	if err != nil {
		jsonError(w, "failed to load activity", http.StatusInternalServerError)
		return
	}

	// Filter out bingo entries (bingo is secret)
	entries := make([]models.ActivityEntry, 0, len(all))
	for _, e := range all {
		if e.Type == "bingo_resolved" || e.Type == "bingo_winner" {
			continue
		}
		entries = append(entries, e)
		if len(entries) >= 50 {
			break
		}
	}
	jsonResp(w, entries, http.StatusOK)
}
