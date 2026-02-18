package handlers

import (
	"encoding/json"
	"net/http"
	"pauls-bach/market"
	"pauls-bach/models"
	"pauls-bach/store"
)

func jsonResp(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func snapshotOdds(s *store.Store, eventID int, odds []market.OutcomeOdds) {
	for _, o := range odds {
		s.OddsSnapshots.Create(&models.OddsSnapshot{
			EventID:   eventID,
			OutcomeID: o.OutcomeID,
			Odds:      o.Odds,
		})
	}
}
