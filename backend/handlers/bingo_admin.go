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

	jsonResp(w, map[string]string{"message": "bingo event resolved"}, http.StatusOK)
}

func (h *BingoAdminHandler) ResolveCustomSquare(w http.ResponseWriter, r *http.Request) {
	boardID, err := strconv.Atoi(chi.URLParam(r, "boardID"))
	if err != nil {
		jsonError(w, "invalid board id", http.StatusBadRequest)
		return
	}
	position, err := strconv.Atoi(chi.URLParam(r, "position"))
	if err != nil || position < 0 || position > 24 {
		jsonError(w, "invalid position", http.StatusBadRequest)
		return
	}
	if !customPositions[position] {
		jsonError(w, "position is not a custom square", http.StatusBadRequest)
		return
	}

	store.WriteLock()
	defer store.WriteUnlock()

	board, err := h.Store.BingoBoards.GetByID(boardID)
	if err != nil {
		jsonError(w, "board not found", http.StatusNotFound)
		return
	}

	for i, sq := range board.Squares {
		if sq.Position == position {
			if sq.Resolved {
				jsonError(w, "square already resolved", http.StatusBadRequest)
				return
			}
			board.Squares[i].Resolved = true
			break
		}
	}

	if err := h.Store.BingoBoards.Update(board); err != nil {
		jsonError(w, "failed to update board", http.StatusInternalServerError)
		return
	}

	h.checkBingo(board)

	h.Broker.Broadcast(sse.EventBingoResolved, map[string]interface{}{
		"board_id": boardID,
		"position": position,
	})

	jsonResp(w, map[string]string{"message": "custom square resolved"}, http.StatusOK)
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
		}
	}
}
