package handlers

import (
	"encoding/json"
	"fmt"
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

	// Log activity
	entry := &models.ActivityEntry{
		Type:    "event_created",
		Message: fmt.Sprintf("New market: '%s'", event.Title),
		EventID: event.ID,
	}
	h.Store.Activity.Create(entry)
	h.Broker.Broadcast(sse.EventActivityNew, entry)

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

func (h *AdminHandler) SetBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req struct {
		Balance int `json:"balance"`
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

	user.Balance = req.Balance
	if err := h.Store.Users.Update(user); err != nil {
		jsonError(w, "failed to update user", http.StatusInternalServerError)
		return
	}

	jsonResp(w, map[string]interface{}{"message": "updated", "balance": user.Balance}, http.StatusOK)
}

func (h *AdminHandler) ResetBingoBoard(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	store.WriteLock()
	defer store.WriteUnlock()

	if err := h.Store.BingoBoards.DeleteByUserID(userID); err != nil {
		jsonError(w, "failed to reset board", http.StatusInternalServerError)
		return
	}
	if err := h.Store.BingoWinners.DeleteByUserID(userID); err != nil {
		jsonError(w, "failed to reset winners", http.StatusInternalServerError)
		return
	}

	jsonResp(w, map[string]string{"message": "bingo board reset"}, http.StatusOK)
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
		Balance  int    `json:"balance"`
	}

	result := make([]userInfo, 0, len(users))
	for _, u := range users {
		result = append(result, userInfo{
			ID:       u.ID,
			Username: u.Username,
			IsAdmin:  u.IsAdmin,
			Bingo:    u.Bingo,
			Balance:  u.Balance,
		})
	}

	jsonResp(w, result, http.StatusOK)
}

type updateEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Outcomes    []struct {
		ID    int    `json:"id"`
		Label string `json:"label"`
	} `json:"outcomes"`
}

func (h *AdminHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	var req updateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Title == "" {
		jsonError(w, "title is required", http.StatusBadRequest)
		return
	}

	store.WriteLock()
	defer store.WriteUnlock()

	event, err := h.Store.Events.GetByID(eventID)
	if err != nil {
		jsonError(w, "event not found", http.StatusNotFound)
		return
	}

	event.Title = req.Title
	event.Description = req.Description
	if err := h.Store.Events.Update(event); err != nil {
		jsonError(w, "failed to update event", http.StatusInternalServerError)
		return
	}

	// Update outcomes: delete old, create new
	if len(req.Outcomes) > 0 {
		if err := h.Store.Outcomes.DeleteByEventID(eventID); err != nil {
			jsonError(w, "failed to update outcomes", http.StatusInternalServerError)
			return
		}
		for _, o := range req.Outcomes {
			outcome := &models.Outcome{EventID: eventID, Label: o.Label}
			if err := h.Store.Outcomes.Create(outcome); err != nil {
				jsonError(w, "failed to create outcome", http.StatusInternalServerError)
				return
			}
		}
	}

	h.Broker.Broadcast(sse.EventEventCreated, map[string]interface{}{
		"event_id": event.ID,
		"title":    event.Title,
	})

	jsonResp(w, map[string]string{"message": "event updated"}, http.StatusOK)
}

func (h *AdminHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	store.WriteLock()
	defer store.WriteUnlock()

	if _, err := h.Store.Events.GetByID(eventID); err != nil {
		jsonError(w, "event not found", http.StatusNotFound)
		return
	}

	// Refund any open positions
	positions, _ := h.Store.Positions.GetByEventID(eventID)
	for _, p := range positions {
		user, err := h.Store.Users.GetByID(p.UserID)
		if err != nil {
			continue
		}
		refund := int(p.Shares)
		user.Balance += refund
		h.Store.Users.Update(user)
	}

	h.Store.Positions.DeleteByEventID(eventID)
	h.Store.Outcomes.DeleteByEventID(eventID)
	h.Store.Transactions.DeleteByEventID(eventID)
	h.Store.OddsSnapshots.DeleteByEventID(eventID)
	if err := h.Store.Events.Delete(eventID); err != nil {
		jsonError(w, "failed to delete event", http.StatusInternalServerError)
		return
	}

	h.Broker.Broadcast(sse.EventEventResolved, map[string]interface{}{
		"event_id": eventID,
		"deleted":  true,
	})

	jsonResp(w, map[string]string{"message": "event deleted"}, http.StatusOK)
}

func (h *AdminHandler) UnresolveEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	store.WriteLock()
	defer store.WriteUnlock()

	event, err := h.Store.Events.GetByID(eventID)
	if err != nil {
		jsonError(w, "event not found", http.StatusNotFound)
		return
	}
	if event.Status != "resolved" {
		jsonError(w, "event is not resolved", http.StatusBadRequest)
		return
	}

	event.Status = "open"
	event.WinningOutcomeID = 0
	event.ResolvedAt = ""
	if err := h.Store.Events.Update(event); err != nil {
		jsonError(w, "failed to unresolve event", http.StatusInternalServerError)
		return
	}

	h.Broker.Broadcast(sse.EventEventCreated, map[string]interface{}{
		"event_id": eventID,
		"title":    event.Title,
	})

	jsonResp(w, map[string]string{"message": "event unresolved"}, http.StatusOK)
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

	// Log activity: resolution
	resolveEntry := &models.ActivityEntry{
		Type:    "event_resolved",
		Message: fmt.Sprintf("'%s' resolved — %s wins!", title, winnerLabel),
		EventID: eventID,
	}
	h.Store.Activity.Create(resolveEntry)
	h.Broker.Broadcast(sse.EventActivityNew, resolveEntry)

	// Log activity: each winner payout
	for _, uo := range result.UserOutcomes {
		if !uo.Won || uo.Payout == 0 {
			continue
		}
		winner, _ := h.Store.Users.GetByID(uo.UserID)
		if winner == nil {
			continue
		}
		payoutEntry := &models.ActivityEntry{
			Type:    "payout",
			Message: fmt.Sprintf("%s won %d pts from '%s'", winner.Username, uo.Payout, title),
			UserID:  uo.UserID,
			EventID: eventID,
		}
		h.Store.Activity.Create(payoutEntry)
		h.Broker.Broadcast(sse.EventActivityNew, payoutEntry)
	}

	jsonResp(w, map[string]string{"message": "event resolved"}, http.StatusOK)
}
