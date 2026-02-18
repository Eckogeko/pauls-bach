package handlers

import (
	"net/http"
	"pauls-bach/market"
	"pauls-bach/middleware"
	"pauls-bach/models"
	"pauls-bach/store"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type EventHandler struct {
	Store  *store.Store
	Engine *market.Engine
}

type eventResponse struct {
	ID               int                `json:"id"`
	Title            string             `json:"title"`
	Description      string             `json:"description"`
	EventType        string             `json:"event_type"`
	Status           string             `json:"status"`
	WinningOutcomeID int                `json:"winning_outcome_id,omitempty"`
	CreatedAt        string             `json:"created_at"`
	ResolvedAt       string             `json:"resolved_at,omitempty"`
	LastTradeAt      string             `json:"last_trade_at,omitempty"`
	Odds             []market.OutcomeOdds `json:"odds"`
}

type eventDetailResponse struct {
	eventResponse
	UserPositions []userPosition `json:"user_positions,omitempty"`
}

type userPosition struct {
	OutcomeID   int     `json:"outcome_id"`
	OutcomeLabel string `json:"outcome_label"`
	Shares      float64 `json:"shares"`
	AvgPrice    float64 `json:"avg_price"`
}

func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	store.ReadLock()
	defer store.ReadUnlock()

	events, err := h.Store.Events.GetAll()
	if err != nil {
		jsonError(w, "failed to load events", http.StatusInternalServerError)
		return
	}

	lastTrades, _ := h.Store.OddsSnapshots.LastSnapshotTimeByEvent()

	var resp []eventResponse
	for _, e := range events {
		odds, _ := h.Engine.GetOdds(e.ID)
		if odds == nil {
			odds = []market.OutcomeOdds{}
		}
		resp = append(resp, eventResponse{
			ID:               e.ID,
			Title:            e.Title,
			Description:      e.Description,
			EventType:        e.EventType,
			Status:           e.Status,
			WinningOutcomeID: e.WinningOutcomeID,
			CreatedAt:        e.CreatedAt,
			ResolvedAt:       e.ResolvedAt,
			LastTradeAt:      lastTrades[e.ID],
			Odds:             odds,
		})
	}
	if resp == nil {
		resp = []eventResponse{}
	}

	sort.Slice(resp, func(i, j int) bool {
		ti := resp[i].LastTradeAt
		if ti == "" {
			ti = resp[i].CreatedAt
		}
		tj := resp[j].LastTradeAt
		if tj == "" {
			tj = resp[j].CreatedAt
		}
		return ti > tj
	})

	jsonResp(w, resp, http.StatusOK)
}

func (h *EventHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, err := strconv.Atoi(idStr)
	if err != nil {
		jsonError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	store.ReadLock()
	defer store.ReadUnlock()

	event, err := h.Store.Events.GetByID(eventID)
	if err != nil {
		jsonError(w, "event not found", http.StatusNotFound)
		return
	}

	odds, _ := h.Engine.GetOdds(eventID)
	if odds == nil {
		odds = []market.OutcomeOdds{}
	}

	resp := eventDetailResponse{
		eventResponse: eventResponse{
			ID:               event.ID,
			Title:            event.Title,
			Description:      event.Description,
			EventType:        event.EventType,
			Status:           event.Status,
			WinningOutcomeID: event.WinningOutcomeID,
			CreatedAt:        event.CreatedAt,
			ResolvedAt:       event.ResolvedAt,
			Odds:             odds,
		},
	}

	// Add user positions if authenticated
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if ok {
		positions, _ := h.Store.Positions.GetByUserAndEvent(userID, eventID)
		for _, p := range positions {
			label := ""
			for _, o := range odds {
				if o.OutcomeID == p.OutcomeID {
					label = o.Label
					break
				}
			}
			resp.UserPositions = append(resp.UserPositions, userPosition{
				OutcomeID:    p.OutcomeID,
				OutcomeLabel: label,
				Shares:       p.Shares,
				AvgPrice:     p.AvgPrice,
			})
		}
	}

	jsonResp(w, resp, http.StatusOK)
}

func (h *EventHandler) OddsHistory(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	store.ReadLock()
	defer store.ReadUnlock()

	snapshots, err := h.Store.OddsSnapshots.GetByEventID(eventID)
	if err != nil {
		jsonError(w, "failed to load history", http.StatusInternalServerError)
		return
	}
	if snapshots == nil {
		snapshots = []*models.OddsSnapshot{}
	}
	jsonResp(w, snapshots, http.StatusOK)
}
