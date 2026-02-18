package store

import (
	"fmt"
	"pauls-bach/models"
	"strconv"
	"time"
)

type EventStore struct {
	filePath string
}

var eventHeader = []string{"id", "title", "description", "event_type", "status", "winning_outcome_id", "created_at", "resolved_at"}

func (s *EventStore) toRow(e *models.Event) []string {
	winID := ""
	if e.WinningOutcomeID != 0 {
		winID = strconv.Itoa(e.WinningOutcomeID)
	}
	return []string{
		strconv.Itoa(e.ID),
		e.Title,
		e.Description,
		e.EventType,
		e.Status,
		winID,
		e.CreatedAt,
		e.ResolvedAt,
	}
}

func (s *EventStore) fromRow(row []string) (*models.Event, error) {
	id, _ := strconv.Atoi(row[0])
	winID, _ := strconv.Atoi(row[5])
	return &models.Event{
		ID:               id,
		Title:            row[1],
		Description:      row[2],
		EventType:        row[3],
		Status:           row[4],
		WinningOutcomeID: winID,
		CreatedAt:        row[6],
		ResolvedAt:       row[7],
	}, nil
}

func (s *EventStore) GetAll() ([]models.Event, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	events := make([]models.Event, 0, len(rows))
	for _, row := range rows {
		e, err := s.fromRow(row)
		if err != nil {
			continue
		}
		events = append(events, *e)
	}
	return events, nil
}

func (s *EventStore) GetByID(id int) (*models.Event, error) {
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
	return nil, fmt.Errorf("event not found")
}

func (s *EventStore) Create(e *models.Event) error {
	id, err := nextID(s.filePath)
	if err != nil {
		return err
	}
	e.ID = id
	e.CreatedAt = time.Now().Format(time.RFC3339)
	return appendRow(s.filePath, s.toRow(e))
}

func (s *EventStore) Update(e *models.Event) error {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return err
	}
	for i, row := range rows {
		rowID, _ := strconv.Atoi(row[0])
		if rowID == e.ID {
			rows[i] = s.toRow(e)
			return writeAllRows(s.filePath, eventHeader, rows)
		}
	}
	return fmt.Errorf("event not found")
}
