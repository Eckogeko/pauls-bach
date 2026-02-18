package store

import (
	"fmt"
	"pauls-bach/models"
	"strconv"
)

type OutcomeStore struct {
	filePath string
}

func (s *OutcomeStore) toRow(o *models.Outcome) []string {
	return []string{
		strconv.Itoa(o.ID),
		strconv.Itoa(o.EventID),
		o.Label,
	}
}

func (s *OutcomeStore) fromRow(row []string) *models.Outcome {
	id, _ := strconv.Atoi(row[0])
	eventID, _ := strconv.Atoi(row[1])
	return &models.Outcome{
		ID:      id,
		EventID: eventID,
		Label:   row[2],
	}
}

func (s *OutcomeStore) GetByEventID(eventID int) ([]models.Outcome, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	var outcomes []models.Outcome
	for _, row := range rows {
		eid, _ := strconv.Atoi(row[1])
		if eid == eventID {
			outcomes = append(outcomes, *s.fromRow(row))
		}
	}
	return outcomes, nil
}

func (s *OutcomeStore) Create(o *models.Outcome) error {
	id, err := nextID(s.filePath)
	if err != nil {
		return err
	}
	o.ID = id
	return appendRow(s.filePath, s.toRow(o))
}

var outcomeHeader = []string{"id", "event_id", "label"}

func (s *OutcomeStore) Update(o *models.Outcome) error {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return err
	}
	for i, row := range rows {
		rowID, _ := strconv.Atoi(row[0])
		if rowID == o.ID {
			rows[i] = s.toRow(o)
			return writeAllRows(s.filePath, outcomeHeader, rows)
		}
	}
	return fmt.Errorf("outcome not found")
}

func (s *OutcomeStore) DeleteByEventID(eventID int) error {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return err
	}
	var kept [][]string
	for _, row := range rows {
		eid, _ := strconv.Atoi(row[1])
		if eid != eventID {
			kept = append(kept, row)
		}
	}
	return writeAllRows(s.filePath, outcomeHeader, kept)
}

func (s *OutcomeStore) GetByID(id int) (*models.Outcome, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		rowID, _ := strconv.Atoi(row[0])
		if rowID == id {
			return s.fromRow(row), nil
		}
	}
	return nil, nil
}
