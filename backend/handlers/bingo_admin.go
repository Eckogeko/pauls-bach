package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pauls-bach/models"
	"pauls-bach/sse"
	"pauls-bach/store"
	"strconv"

	"github.com/go-chi/chi/v5"
)

var lineLabels = map[string]string{
	"row-0":  "Top Row",
	"row-1":  "2nd Row",
	"row-2":  "Middle Row",
	"row-3":  "4th Row",
	"row-4":  "Bottom Row",
	"col-0":  "1st Column",
	"col-1":  "2nd Column",
	"col-2":  "Middle Column",
	"col-3":  "4th Column",
	"col-4":  "5th Column",
	"diag-0": "Diagonal ↘",
	"diag-1": "Diagonal ↗",
}

func readableLineName(line string) string {
	if label, ok := lineLabels[line]; ok {
		return label
	}
	return line
}

type BingoAdminHandler struct {
	Store  *store.Store
	Broker *sse.Broker
}

type createBingoEventRequest struct {
	Title  string `json:"title"`
	Rarity string `json:"rarity"`
}

func (h *BingoAdminHandler) CreateBingoEvent(w http.ResponseWriter, r *http.Request) {
	var req createBingoEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Title == "" {
		jsonError(w, "title is required", http.StatusBadRequest)
		return
	}
	if req.Rarity == "" {
		req.Rarity = "common"
	}
	if req.Rarity != "common" && req.Rarity != "uncommon" {
		jsonError(w, "rarity must be 'common' or 'uncommon'", http.StatusBadRequest)
		return
	}

	store.WriteLock()
	defer store.WriteUnlock()

	event := &models.BingoEvent{Title: req.Title, Rarity: req.Rarity}
	if err := h.Store.BingoEvents.Create(event); err != nil {
		jsonError(w, "failed to create bingo event", http.StatusInternalServerError)
		return
	}

	jsonResp(w, event, http.StatusCreated)
}

func (h *BingoAdminHandler) UpdateBingoEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	var req createBingoEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Title == "" {
		jsonError(w, "title is required", http.StatusBadRequest)
		return
	}
	if req.Rarity == "" {
		req.Rarity = "common"
	}
	if req.Rarity != "common" && req.Rarity != "uncommon" {
		jsonError(w, "rarity must be 'common' or 'uncommon'", http.StatusBadRequest)
		return
	}

	store.WriteLock()
	defer store.WriteUnlock()

	event, err := h.Store.BingoEvents.GetByID(eventID)
	if err != nil {
		jsonError(w, "bingo event not found", http.StatusNotFound)
		return
	}

	event.Title = req.Title
	event.Rarity = req.Rarity
	if err := h.Store.BingoEvents.Update(event); err != nil {
		jsonError(w, "failed to update bingo event", http.StatusInternalServerError)
		return
	}

	h.Broker.Broadcast(sse.EventBingoResolved, map[string]interface{}{
		"bingo_event_id": eventID,
		"title":          event.Title,
	})

	jsonResp(w, event, http.StatusOK)
}

func (h *BingoAdminHandler) UnresolveBingoEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	store.WriteLock()
	defer store.WriteUnlock()

	event, err := h.Store.BingoEvents.GetByID(eventID)
	if err != nil {
		jsonError(w, "bingo event not found", http.StatusNotFound)
		return
	}
	if !event.Resolved {
		jsonError(w, "event is not resolved", http.StatusBadRequest)
		return
	}

	event.Resolved = false
	if err := h.Store.BingoEvents.Update(event); err != nil {
		jsonError(w, "failed to unresolve event", http.StatusInternalServerError)
		return
	}

	// Update all boards: unmark this event's squares
	boards, _ := h.Store.BingoBoards.GetAll()
	for _, board := range boards {
		changed := false
		for i, sq := range board.Squares {
			if sq.BingoEventID == eventID && sq.Resolved {
				board.Squares[i].Resolved = false
				changed = true
			}
		}
		if changed {
			h.Store.BingoBoards.Update(&board)
		}
	}

	// Remove any bingo winners that depended on this event's squares
	// (simplest: re-check all boards and remove invalid wins)
	for _, board := range boards {
		h.removeInvalidWins(&board)
	}

	h.Broker.Broadcast(sse.EventBingoResolved, map[string]interface{}{
		"bingo_event_id": eventID,
		"title":          event.Title,
		"unresolved":     true,
	})

	jsonResp(w, map[string]string{"message": "bingo event unresolved"}, http.StatusOK)
}

func (h *BingoAdminHandler) ResolveBingoEvent(w http.ResponseWriter, r *http.Request) {
	eventID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid event id", http.StatusBadRequest)
		return
	}

	store.WriteLock()
	defer store.WriteUnlock()

	event, err := h.Store.BingoEvents.GetByID(eventID)
	if err != nil {
		jsonError(w, "bingo event not found", http.StatusNotFound)
		return
	}
	if event.Resolved {
		jsonError(w, "event already resolved", http.StatusBadRequest)
		return
	}

	event.Resolved = true
	if err := h.Store.BingoEvents.Update(event); err != nil {
		jsonError(w, "failed to resolve event", http.StatusInternalServerError)
		return
	}

	// Update all boards containing this event
	boards, _ := h.Store.BingoBoards.GetAll()
	for _, board := range boards {
		changed := false
		for i, sq := range board.Squares {
			if sq.BingoEventID == eventID && !sq.Resolved {
				board.Squares[i].Resolved = true
				changed = true
			}
		}
		if changed {
			h.Store.BingoBoards.Update(&board)
			h.checkBingo(&board)
		}
	}

	h.Broker.Broadcast(sse.EventBingoResolved, map[string]interface{}{
		"bingo_event_id": eventID,
		"title":          event.Title,
	})

	// Log activity
	entry := &models.ActivityEntry{
		Type:    "bingo_resolved",
		Message: fmt.Sprintf("Bingo event resolved: %s", event.Title),
	}
	h.Store.Activity.Create(entry)
	h.Broker.Broadcast(sse.EventActivityNew, entry)

	jsonResp(w, map[string]string{"message": "bingo event resolved"}, http.StatusOK)
}

// 12 possible bingo lines on a 5x5 board
var bingoLines = []struct {
	name      string
	positions [5]int
}{
	{"row-0", [5]int{0, 1, 2, 3, 4}},
	{"row-1", [5]int{5, 6, 7, 8, 9}},
	{"row-2", [5]int{10, 11, 12, 13, 14}},
	{"row-3", [5]int{15, 16, 17, 18, 19}},
	{"row-4", [5]int{20, 21, 22, 23, 24}},
	{"col-0", [5]int{0, 5, 10, 15, 20}},
	{"col-1", [5]int{1, 6, 11, 16, 21}},
	{"col-2", [5]int{2, 7, 12, 17, 22}},
	{"col-3", [5]int{3, 8, 13, 18, 23}},
	{"col-4", [5]int{4, 9, 14, 19, 24}},
	{"diag-0", [5]int{0, 6, 12, 18, 24}},
	{"diag-1", [5]int{4, 8, 12, 16, 20}},
}

func (h *BingoAdminHandler) removeInvalidWins(board *models.BingoBoard) {
	resolved := make(map[int]bool)
	for _, sq := range board.Squares {
		if sq.Resolved {
			resolved[sq.Position] = true
		}
	}

	existingWins, _ := h.Store.BingoWinners.GetByBoardID(board.ID)
	for _, w := range existingWins {
		for _, line := range bingoLines {
			if line.name != w.Line {
				continue
			}
			stillValid := true
			for _, pos := range line.positions {
				if !resolved[pos] {
					stillValid = false
					break
				}
			}
			if !stillValid {
				h.Store.BingoWinners.DeleteByBoardIDAndLine(board.ID, w.Line)
			}
		}
	}
}

func (h *BingoAdminHandler) checkBingo(board *models.BingoBoard) {
	resolved := make(map[int]bool)
	for _, sq := range board.Squares {
		if sq.Resolved {
			resolved[sq.Position] = true
		}
	}

	existingWins, _ := h.Store.BingoWinners.GetByBoardID(board.ID)
	wonLines := make(map[string]bool)
	for _, w := range existingWins {
		wonLines[w.Line] = true
	}

	user, _ := h.Store.Users.GetByID(board.UserID)
	username := ""
	if user != nil {
		username = user.Username
	}

	for _, line := range bingoLines {
		if wonLines[line.name] {
			continue
		}
		allResolved := true
		for _, pos := range line.positions {
			if !resolved[pos] {
				allResolved = false
				break
			}
		}
		if allResolved {
			winner := &models.BingoWinner{
				UserID:   board.UserID,
				Username: username,
				BoardID:  board.ID,
				Line:     line.name,
			}
			h.Store.BingoWinners.Create(winner)

			h.Broker.Broadcast(sse.EventBingoWinner, map[string]interface{}{
				"username": username,
				"line":     line.name,
				"message":  fmt.Sprintf("%s got BINGO! (%s)", username, line.name),
			})

			// Log activity
			bingoEntry := &models.ActivityEntry{
				Type:    "bingo_winner",
				Message: fmt.Sprintf("%s got BINGO! (%s)", username, readableLineName(line.name)),
				UserID:  board.UserID,
			}
			h.Store.Activity.Create(bingoEntry)
			h.Broker.Broadcast(sse.EventActivityNew, bingoEntry)
		}
	}
}
