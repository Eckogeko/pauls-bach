package handlers

import (
	"encoding/json"
	"net/http"
	"pauls-bach/middleware"
	"pauls-bach/models"
	"pauls-bach/store"
)

type BingoHandler struct {
	Store *store.Store
}

// ListBingoEvents returns all bingo events for board building.
func (h *BingoHandler) ListBingoEvents(w http.ResponseWriter, r *http.Request) {
	store.ReadLock()
	defer store.ReadUnlock()

	events, err := h.Store.BingoEvents.GetAll()
	if err != nil {
		jsonError(w, "failed to load bingo events", http.StatusInternalServerError)
		return
	}
	jsonResp(w, events, http.StatusOK)
}

// GetBoard returns the current user's bingo board (or 404).
func (h *BingoHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	store.ReadLock()
	defer store.ReadUnlock()

	board, err := h.Store.BingoBoards.GetByUserID(userID)
	if err != nil {
		jsonError(w, "failed to load board", http.StatusInternalServerError)
		return
	}
	if board == nil {
		jsonError(w, "no board", http.StatusNotFound)
		return
	}

	// Attach winners for this board
	winners, _ := h.Store.BingoWinners.GetByBoardID(board.ID)

	jsonResp(w, map[string]interface{}{
		"board":   board,
		"winners": winners,
	}, http.StatusOK)
}

var customPositions = map[int]bool{0: true, 4: true, 12: true, 20: true, 24: true}

type createBoardRequest struct {
	Squares []models.BingoSquare `json:"squares"`
}

// CreateBoard creates a locked bingo board for the current user.
func (h *BingoHandler) CreateBoard(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var req createBoardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	if len(req.Squares) != 25 {
		jsonError(w, "board must have exactly 25 squares", http.StatusBadRequest)
		return
	}

	// Validate squares
	seen := make(map[int]bool)
	seenEvents := make(map[int]bool)
	for _, sq := range req.Squares {
		if sq.Position < 0 || sq.Position > 24 {
			jsonError(w, "invalid square position", http.StatusBadRequest)
			return
		}
		if seen[sq.Position] {
			jsonError(w, "duplicate position", http.StatusBadRequest)
			return
		}
		seen[sq.Position] = true

		if customPositions[sq.Position] {
			if sq.CustomText == "" {
				jsonError(w, "custom squares require text", http.StatusBadRequest)
				return
			}
		} else {
			if sq.BingoEventID == 0 {
				jsonError(w, "non-custom squares require a bingo event", http.StatusBadRequest)
				return
			}
			if seenEvents[sq.BingoEventID] {
				jsonError(w, "each event can only be used once per board", http.StatusBadRequest)
				return
			}
			seenEvents[sq.BingoEventID] = true
		}
	}

	store.WriteLock()
	defer store.WriteUnlock()

	existing, _ := h.Store.BingoBoards.GetByUserID(userID)
	if existing != nil {
		jsonError(w, "board already exists", http.StatusConflict)
		return
	}

	// Validate that all referenced bingo events exist
	for _, sq := range req.Squares {
		if sq.BingoEventID != 0 {
			if _, err := h.Store.BingoEvents.GetByID(sq.BingoEventID); err != nil {
				jsonError(w, "invalid bingo event ID", http.StatusBadRequest)
				return
			}
		}
	}

	// Mark already-resolved events
	for i, sq := range req.Squares {
		if sq.BingoEventID != 0 {
			ev, _ := h.Store.BingoEvents.GetByID(sq.BingoEventID)
			if ev != nil && ev.Resolved {
				req.Squares[i].Resolved = true
			}
		}
	}

	board := &models.BingoBoard{
		UserID:  userID,
		Squares: req.Squares,
	}
	if err := h.Store.BingoBoards.Create(board); err != nil {
		jsonError(w, "failed to create board", http.StatusInternalServerError)
		return
	}

	jsonResp(w, board, http.StatusCreated)
}

// ListWinners returns all bingo winners.
func (h *BingoHandler) ListWinners(w http.ResponseWriter, r *http.Request) {
	store.ReadLock()
	defer store.ReadUnlock()

	winners, err := h.Store.BingoWinners.GetAll()
	if err != nil {
		jsonError(w, "failed to load winners", http.StatusInternalServerError)
		return
	}
	jsonResp(w, winners, http.StatusOK)
}
