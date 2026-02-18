package store

import (
	"encoding/json"
	"fmt"
	"pauls-bach/models"
	"strconv"
	"time"
)

type BingoBoardStore struct {
	filePath string
}

var bingoBoardHeader = []string{"id", "user_id", "squares", "created_at"}

func (s *BingoBoardStore) toRow(b *models.BingoBoard) []string {
	sq, _ := json.Marshal(b.Squares)
	return []string{
		strconv.Itoa(b.ID),
		strconv.Itoa(b.UserID),
		string(sq),
		b.CreatedAt,
	}
}

func (s *BingoBoardStore) fromRow(row []string) (*models.BingoBoard, error) {
	id, _ := strconv.Atoi(row[0])
	userID, _ := strconv.Atoi(row[1])
	var squares []models.BingoSquare
	if err := json.Unmarshal([]byte(row[2]), &squares); err != nil {
		return nil, err
	}
	return &models.BingoBoard{
		ID:        id,
		UserID:    userID,
		Squares:   squares,
		CreatedAt: row[3],
	}, nil
}

func (s *BingoBoardStore) GetAll() ([]models.BingoBoard, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	boards := make([]models.BingoBoard, 0, len(rows))
	for _, row := range rows {
		b, err := s.fromRow(row)
		if err != nil {
			continue
		}
		boards = append(boards, *b)
	}
	return boards, nil
}

func (s *BingoBoardStore) GetByID(id int) (*models.BingoBoard, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		rowID, _ := strconv.Atoi(row[0])
		if rowID == id {
			return s.fromRow(row)
		}
	}
	return nil, fmt.Errorf("bingo board not found")
}

func (s *BingoBoardStore) GetByUserID(userID int) (*models.BingoBoard, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		uid, _ := strconv.Atoi(row[1])
		if uid == userID {
			return s.fromRow(row)
		}
	}
	return nil, nil // no board yet
}

func (s *BingoBoardStore) Create(b *models.BingoBoard) error {
	id, err := nextID(s.filePath)
	if err != nil {
		return err
	}
	b.ID = id
	b.CreatedAt = time.Now().Format(time.RFC3339)
	return appendRow(s.filePath, s.toRow(b))
}

func (s *BingoBoardStore) Update(b *models.BingoBoard) error {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return err
	}
	for i, row := range rows {
		rowID, _ := strconv.Atoi(row[0])
		if rowID == b.ID {
			rows[i] = s.toRow(b)
			return writeAllRows(s.filePath, bingoBoardHeader, rows)
		}
	}
	return fmt.Errorf("bingo board not found")
}

func (s *BingoBoardStore) DeleteByUserID(userID int) error {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return err
	}
	var kept [][]string
	for _, row := range rows {
		uid, _ := strconv.Atoi(row[1])
		if uid != userID {
			kept = append(kept, row)
		}
	}
	return writeAllRows(s.filePath, bingoBoardHeader, kept)
}
