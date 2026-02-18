package handlers

import (
	"net/http"
	"pauls-bach/middleware"
	"pauls-bach/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type HistoryHandler struct {
	Store *store.Store
}

type historyEntry struct {
	ID            int     `json:"id"`
	EventID       int     `json:"event_id"`
	EventTitle    string  `json:"event_title"`
	OutcomeID     int     `json:"outcome_id"`
	OutcomeLabel  string  `json:"outcome_label"`
	TxType        string  `json:"tx_type"`
	Shares        float64 `json:"shares"`
	Points        int     `json:"points"`
	CreatedAt     string  `json:"created_at"`
}

func (h *HistoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	requestedID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	// Users can only see their own history unless admin
	callerID := r.Context().Value(middleware.UserIDKey).(int)
	isAdmin, _ := r.Context().Value(middleware.IsAdminKey).(bool)
	if callerID != requestedID && !isAdmin {
		jsonError(w, "forbidden", http.StatusForbidden)
		return
	}

	store.ReadLock()
	defer store.ReadUnlock()

	txs, err := h.Store.Transactions.GetByUserID(requestedID)
	if err != nil {
		jsonError(w, "failed to load history", http.StatusInternalServerError)
		return
	}

	// Build event/outcome name cache
	events, _ := h.Store.Events.GetAll()
	eventMap := make(map[int]string)
	for _, e := range events {
		eventMap[e.ID] = e.Title
	}

	entries := make([]historyEntry, 0, len(txs))
	for i := len(txs) - 1; i >= 0; i-- { // newest first
		tx := txs[i]
		outcomeLabel := ""
		if outcomes, _ := h.Store.Outcomes.GetByEventID(tx.EventID); outcomes != nil {
			for _, o := range outcomes {
				if o.ID == tx.OutcomeID {
					outcomeLabel = o.Label
					break
				}
			}
		}
		entries = append(entries, historyEntry{
			ID:           tx.ID,
			EventID:      tx.EventID,
			EventTitle:   eventMap[tx.EventID],
			OutcomeID:    tx.OutcomeID,
			OutcomeLabel: outcomeLabel,
			TxType:       tx.TxType,
			Shares:       tx.Shares,
			Points:       tx.Points,
			CreatedAt:    tx.CreatedAt,
		})
	}

	jsonResp(w, entries, http.StatusOK)
}
