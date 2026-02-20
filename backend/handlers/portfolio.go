package handlers

import (
	"math"
	"net/http"
	"pauls-bach/market"
	"pauls-bach/middleware"
	"pauls-bach/store"
)

type PortfolioHandler struct {
	Store  *store.Store
	Engine *market.Engine
}

type portfolioPosition struct {
	EventID        int     `json:"event_id"`
	EventTitle     string  `json:"event_title"`
	OutcomeID      int     `json:"outcome_id"`
	OutcomeLabel   string  `json:"outcome_label"`
	Shares         float64 `json:"shares"`
	AvgPrice       float64 `json:"avg_price"`
	PotentialPayout int    `json:"potential_payout"`
}

type portfolioResponse struct {
	Positions      []portfolioPosition `json:"positions"`
	TotalInvested  int                 `json:"total_invested"`
	TotalPotential int                 `json:"total_potential"`
	ActiveMarkets  int                 `json:"active_markets"`
}

func (h *PortfolioHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	store.ReadLock()
	defer store.ReadUnlock()

	positions, err := h.Store.Positions.GetByUserID(userID)
	if err != nil {
		jsonError(w, "failed to load positions", http.StatusInternalServerError)
		return
	}

	resp := portfolioResponse{
		Positions: make([]portfolioPosition, 0),
	}
	eventsSeen := make(map[int]bool)

	for _, p := range positions {
		event, err := h.Store.Events.GetByID(p.EventID)
		if err != nil || event.Status == "resolved" {
			continue
		}

		odds, _ := h.Engine.GetOdds(p.EventID)
		outcomeLabel := ""
		var totalPool float64
		var outcomeShares float64
		for _, o := range odds {
			totalPool += o.Shares
			if o.OutcomeID == p.OutcomeID {
				outcomeLabel = o.Label
				outcomeShares = o.Shares
			}
		}

		poolPayout := 0
		if outcomeShares > 0 {
			poolPayout = int(math.Round(totalPool * (p.Shares / outcomeShares)))
		}
		bonus := int(math.Round(p.Shares*0.25)) + 20
		potentialPayout := poolPayout + bonus

		resp.Positions = append(resp.Positions, portfolioPosition{
			EventID:        p.EventID,
			EventTitle:     event.Title,
			OutcomeID:      p.OutcomeID,
			OutcomeLabel:   outcomeLabel,
			Shares:         p.Shares,
			AvgPrice:       p.AvgPrice,
			PotentialPayout: potentialPayout,
		})

		resp.TotalInvested += int(math.Round(p.Shares))
		resp.TotalPotential += potentialPayout
		if !eventsSeen[p.EventID] {
			eventsSeen[p.EventID] = true
			resp.ActiveMarkets++
		}
	}

	jsonResp(w, resp, http.StatusOK)
}
