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

	entries, err := h.Store.Activity.GetRecent(50)
	if err != nil {
		jsonError(w, "failed to load activity", http.StatusInternalServerError)
		return
	}
	if entries == nil {
		entries = []models.ActivityEntry{}
	}
	jsonResp(w, entries, http.StatusOK)
}
