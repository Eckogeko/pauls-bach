package store

import (
	"pauls-bach/models"
	"strconv"
	"time"
)

type BingoWinnerStore struct {
	filePath string
}

var bingoWinnerHeader = []string{"id", "user_id", "username", "board_id", "line", "created_at"}

func (s *BingoWinnerStore) toRow(w *models.BingoWinner) []string {
	return []string{
		strconv.Itoa(w.ID),
		strconv.Itoa(w.UserID),
		w.Username,
		strconv.Itoa(w.BoardID),
		w.Line,
		w.CreatedAt,
	}
}

func (s *BingoWinnerStore) fromRow(row []string) (*models.BingoWinner, error) {
	id, _ := strconv.Atoi(row[0])
	userID, _ := strconv.Atoi(row[1])
	boardID, _ := strconv.Atoi(row[3])
	return &models.BingoWinner{
		ID:        id,
		UserID:    userID,
		Username:  row[2],
		BoardID:   boardID,
		Line:      row[4],
		CreatedAt: row[5],
	}, nil
}

func (s *BingoWinnerStore) GetAll() ([]models.BingoWinner, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	winners := make([]models.BingoWinner, 0, len(rows))
	for _, row := range rows {
		w, err := s.fromRow(row)
		if err != nil {
			continue
		}
		winners = append(winners, *w)
	}
	return winners, nil
}

func (s *BingoWinnerStore) GetByBoardID(boardID int) ([]models.BingoWinner, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	var winners []models.BingoWinner
	for _, row := range rows {
		bid, _ := strconv.Atoi(row[3])
		if bid == boardID {
			w, err := s.fromRow(row)
			if err != nil {
				continue
			}
			winners = append(winners, *w)
		}
	}
	return winners, nil
}

func (s *BingoWinnerStore) Create(w *models.BingoWinner) error {
	id, err := nextID(s.filePath)
	if err != nil {
		return err
	}
	w.ID = id
	w.CreatedAt = time.Now().Format(time.RFC3339)
	return appendRow(s.filePath, s.toRow(w))
}
