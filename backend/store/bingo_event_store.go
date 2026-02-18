package store

import (
	"fmt"
	"pauls-bach/models"
	"strconv"
	"time"
)

type BingoEventStore struct {
	filePath string
}

var bingoEventHeader = []string{"id", "title", "rarity", "resolved", "created_at"}

func (s *BingoEventStore) toRow(e *models.BingoEvent) []string {
	return []string{
		strconv.Itoa(e.ID),
		e.Title,
		e.Rarity,
		strconv.FormatBool(e.Resolved),
		e.CreatedAt,
	}
}

func (s *BingoEventStore) fromRow(row []string) (*models.BingoEvent, error) {
	id, _ := strconv.Atoi(row[0])
	// Backward compat: old rows without rarity column
	if len(row) == 4 {
		return &models.BingoEvent{
			ID:        id,
			Title:     row[1],
			Rarity:    "common",
			Resolved:  row[2] == "true",
			CreatedAt: row[3],
		}, nil
	}
	return &models.BingoEvent{
		ID:        id,
		Title:     row[1],
		Rarity:    row[2],
		Resolved:  row[3] == "true",
		CreatedAt: row[4],
	}, nil
}

func (s *BingoEventStore) GetAll() ([]models.BingoEvent, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	events := make([]models.BingoEvent, 0, len(rows))
	for _, row := range rows {
		e, err := s.fromRow(row)
		if err != nil {
			continue
		}
		events = append(events, *e)
	}
	return events, nil
}

func (s *BingoEventStore) GetByID(id int) (*models.BingoEvent, error) {
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
	return nil, fmt.Errorf("bingo event not found")
}

func (s *BingoEventStore) Create(e *models.BingoEvent) error {
	id, err := nextID(s.filePath)
	if err != nil {
		return err
	}
	e.ID = id
	e.CreatedAt = time.Now().Format(time.RFC3339)
	return appendRow(s.filePath, s.toRow(e))
}

func (s *BingoEventStore) Update(e *models.BingoEvent) error {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return err
	}
	for i, row := range rows {
		rowID, _ := strconv.Atoi(row[0])
		if rowID == e.ID {
			rows[i] = s.toRow(e)
			return writeAllRows(s.filePath, bingoEventHeader, rows)
		}
	}
	return fmt.Errorf("bingo event not found")
}
