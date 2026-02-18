package handlers

import (
	"encoding/json"
	"net/http"
	"pauls-bach/market"
	"pauls-bach/middleware"
	"pauls-bach/sse"
	"pauls-bach/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type TradingHandler struct {
	Store  *store.Store
	Engine *market.Engine
	Broker *sse.Broker
}

type buyRequest struct {
	OutcomeID int `json:"outcome_id"`
	Amount    int `json:"amount"`
}

type sellRequest struct {
	OutcomeID int     `json:"outcome_id"`
	Shares    float64 `json:"shares"`
}

func (h *TradingHandler) Buy(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	eventID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	var req buyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	store.WriteLock()
	defer store.WriteUnlock()

	if err := h.Engine.Buy(userID, eventID, req.OutcomeID, req.Amount); err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	odds, _ := h.Engine.GetOdds(eventID)
	user, _ := h.Store.Users.GetByID(userID)
	snapshotOdds(h.Store, eventID, odds)

	h.Broker.Broadcast(sse.EventOddsUpdated, map[string]interface{}{
		"event_id": eventID,
		"odds":     odds,
	})

	jsonResp(w, map[string]interface{}{
		"message": "purchase successful",
		"odds":    odds,
		"balance": user.Balance,
	}, http.StatusOK)
}

func (h *TradingHandler) Sell(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	eventID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	var req sellRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	store.WriteLock()
	defer store.WriteUnlock()

	pointsBack, err := h.Engine.Sell(userID, eventID, req.OutcomeID, req.Shares)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	odds, _ := h.Engine.GetOdds(eventID)
	user, _ := h.Store.Users.GetByID(userID)
	snapshotOdds(h.Store, eventID, odds)

	h.Broker.Broadcast(sse.EventOddsUpdated, map[string]interface{}{
		"event_id": eventID,
		"odds":     odds,
	})

	jsonResp(w, map[string]interface{}{
		"message":     "sale successful",
		"points_back": pointsBack,
		"odds":        odds,
		"balance":     user.Balance,
	}, http.StatusOK)
}
