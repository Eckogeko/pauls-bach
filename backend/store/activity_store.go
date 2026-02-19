package store

import (
	"pauls-bach/models"
	"strconv"
	"time"
)

type ActivityStore struct {
	filePath string
}

func (s *ActivityStore) toRow(e *models.ActivityEntry) []string {
	return []string{
		strconv.Itoa(e.ID),
		e.Type,
		e.Message,
		strconv.Itoa(e.UserID),
		strconv.Itoa(e.EventID),
		e.CreatedAt,
	}
}

func (s *ActivityStore) fromRow(row []string) *models.ActivityEntry {
	id, _ := strconv.Atoi(row[0])
	userID, _ := strconv.Atoi(row[3])
	eventID, _ := strconv.Atoi(row[4])
	return &models.ActivityEntry{
		ID:        id,
		Type:      row[1],
		Message:   row[2],
		UserID:    userID,
		EventID:   eventID,
		CreatedAt: row[5],
	}
}

func (s *ActivityStore) Create(e *models.ActivityEntry) error {
	id, err := nextID(s.filePath)
	if err != nil {
		return err
	}
	e.ID = id
	e.CreatedAt = time.Now().Format(time.RFC3339)
	return appendRow(s.filePath, s.toRow(e))
}

func (s *ActivityStore) GetRecent(limit int) ([]models.ActivityEntry, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}

	// Return last N entries, newest first
	var entries []models.ActivityEntry
	start := len(rows) - limit
	if start < 0 {
		start = 0
	}
	for i := len(rows) - 1; i >= start; i-- {
		entries = append(entries, *s.fromRow(rows[i]))
	}
	return entries, nil
}
