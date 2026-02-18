package handlers

import (
	"encoding/json"
	"net/http"
	"pauls-bach/market"
	"pauls-bach/models"
	"pauls-bach/sse"
	"pauls-bach/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type AdminHandler struct {
	Store  *store.Store
	Engine *market.Engine
	Broker *sse.Broker
}

type createEventRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	EventType   string   `json:"event_type"`
	Outcomes    []string `json:"outcomes"` // For multi; ignored for binary
}

type resolveRequest struct {
	WinningOutcomeID int `json:"winning_outcome_id"`
}

func (h *AdminHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req createEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Title == "" {
		jsonError(w, "title is required", http.StatusBadRequest)
		return
	}
	if req.EventType != "binary" && req.EventType != "multi" {
		jsonError(w, "event_type must be 'binary' or 'multi'", http.StatusBadRequest)
		return
	}
	if req.EventType == "multi" && len(req.Outcomes) < 2 {
		jsonError(w, "multi events need at least 2 outcomes", http.StatusBadRequest)
		return
	}

	store.WriteLock()
	defer store.WriteUnlock()

	event := &models.Event{
		Title:       req.Title,
		Description: req.Description,
		EventType:   req.EventType,
		Status:      "open",
	}
	if err := h.Store.Events.Create(event); err != nil {
		jsonError(w, "failed to create event", http.StatusInternalServerError)
		return
	}

	var outcomeLabels []string
	if req.EventType == "binary" {
		outcomeLabels = []string{"Yes", "No"}
	} else {
		outcomeLabels = req.Outcomes
	}

	for _, label := range outcomeLabels {
		outcome := &models.Outcome{
			EventID: event.ID,
			Label:   label,
		}
		if err := h.Store.Outcomes.Create(outcome); err != nil {
			jsonError(w, "failed to create outcome", http.StatusInternalServerError)
			return
		}
	}

	odds, _ := h.Engine.GetOdds(event.ID)
	snapshotOdds(h.Store, event.ID, odds)

	h.Broker.Broadcast(sse.EventEventCreated, map[string]interface{}{
		"event_id":    event.ID,
		"title":       event.Title,
		"description": event.Description,
		"event_type":  event.EventType,
		"odds":        odds,
	})

	jsonResp(w, map[string]interface{}{
		"event": event,
		"odds":  odds,
	}, http.StatusCreated)
}

func (h *AdminHandler) SetBingo(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req struct {
		Bingo bool `json:"bingo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	store.WriteLock()
	defer store.WriteUnlock()

	user, err := h.Store.Users.GetByID(userID)
	if err != nil {
		jsonError(w, "user not found", http.StatusNotFound)
		return
	}

	user.Bingo = req.Bingo
	if err := h.Store.Users.Update(user); err != nil {
		jsonError(w, "failed to update user", http.StatusInternalServerError)
		return
	}

	jsonResp(w, map[string]interface{}{"message": "updated", "bingo": user.Bingo}, http.StatusOK)
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	store.ReadLock()
	defer store.ReadUnlock()

	users, err := h.Store.Users.GetAll()
	if err != nil {
		jsonError(w, "failed to load users", http.StatusInternalServerError)
		return
	}

	type userInfo struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		IsAdmin  bool   `json:"is_admin"`
		Bingo    bool   `json:"bingo"`
	}

	result := make([]userInfo, 0, len(users))
	for _, u := range users {
		result = append(result, userInfo{
			ID:       u.ID,
			Username: u.Username,
			IsAdmin:  u.IsAdmin,
			Bingo:    u.Bingo,
		})
	}

	jsonResp(w, result, http.StatusOK)
}

func (h *AdminHandler) ResolveEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	var req resolveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	store.WriteLock()
	defer store.WriteUnlock()

	// Get event title before resolving (positions get cleaned up during resolve)
	event, _ := h.Store.Events.GetByID(eventID)

	result, err := h.Engine.Resolve(eventID, req.WinningOutcomeID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get winner label for broadcast
	winnerLabel := ""
	if event != nil {
		odds, _ := h.Engine.GetOdds(eventID)
		for _, o := range odds {
			if o.OutcomeID == req.WinningOutcomeID {
				winnerLabel = o.Label
				break
			}
		}
	}

	// Send personalized notifications to users who held positions
	title := ""
	if event != nil {
		title = event.Title
	}
	for _, uo := range result.UserOutcomes {
		h.Broker.Send(uo.UserID, sse.EventUserResolved, map[string]interface{}{
			"won":    uo.Won,
			"payout": uo.Payout,
			"refund": uo.Refund,
			"title":  title,
		})
	}

	// Broadcast generic notification for users without positions
	h.Broker.Broadcast(sse.EventEventResolved, map[string]interface{}{
		"event_id":           eventID,
		"title":              title,
		"winning_outcome_id": req.WinningOutcomeID,
		"winner_label":       winnerLabel,
	})

	jsonResp(w, map[string]string{"message": "event resolved"}, http.StatusOK)
}
