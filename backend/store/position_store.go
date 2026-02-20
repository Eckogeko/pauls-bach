package store

import (
	"fmt"
	"pauls-bach/models"
	"strconv"
	"time"
)

type PositionStore struct {
	filePath string
}

var positionHeader = []string{"id", "user_id", "event_id", "outcome_id", "shares", "avg_price", "created_at"}

func (s *PositionStore) toRow(p *models.Position) []string {
	return []string{
		strconv.Itoa(p.ID),
		strconv.Itoa(p.UserID),
		strconv.Itoa(p.EventID),
		strconv.Itoa(p.OutcomeID),
		strconv.FormatFloat(p.Shares, 'f', 6, 64),
		strconv.FormatFloat(p.AvgPrice, 'f', 6, 64),
		p.CreatedAt,
	}
}

func (s *PositionStore) fromRow(row []string) *models.Position {
	id, _ := strconv.Atoi(row[0])
	userID, _ := strconv.Atoi(row[1])
	eventID, _ := strconv.Atoi(row[2])
	outcomeID, _ := strconv.Atoi(row[3])
	shares, _ := strconv.ParseFloat(row[4], 64)
	avgPrice, _ := strconv.ParseFloat(row[5], 64)
	return &models.Position{
		ID:        id,
		UserID:    userID,
		EventID:   eventID,
		OutcomeID: outcomeID,
		Shares:    shares,
		AvgPrice:  avgPrice,
		CreatedAt: row[6],
	}
}

func (s *PositionStore) GetByUserID(userID int) ([]models.Position, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	var positions []models.Position
	for _, row := range rows {
		uid, _ := strconv.Atoi(row[1])
		if uid == userID {
			positions = append(positions, *s.fromRow(row))
		}
	}
	return positions, nil
}

func (s *PositionStore) GetByEventID(eventID int) ([]models.Position, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	var positions []models.Position
	for _, row := range rows {
		eid, _ := strconv.Atoi(row[2])
		if eid == eventID {
			positions = append(positions, *s.fromRow(row))
		}
	}
	return positions, nil
}

func (s *PositionStore) GetByUserAndEvent(userID, eventID int) ([]models.Position, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	var positions []models.Position
	for _, row := range rows {
		uid, _ := strconv.Atoi(row[1])
		eid, _ := strconv.Atoi(row[2])
		if uid == userID && eid == eventID {
			positions = append(positions, *s.fromRow(row))
		}
	}
	return positions, nil
}

func (s *PositionStore) GetByUserEventOutcome(userID, eventID, outcomeID int) (*models.Position, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		uid, _ := strconv.Atoi(row[1])
		eid, _ := strconv.Atoi(row[2])
		oid, _ := strconv.Atoi(row[3])
		if uid == userID && eid == eventID && oid == outcomeID {
			return s.fromRow(row), nil
		}
	}
	return nil, nil
}

func (s *PositionStore) Create(p *models.Position) error {
	id, err := nextID(s.filePath)
	if err != nil {
		return err
	}
	p.ID = id
	p.CreatedAt = time.Now().Format(time.RFC3339)
	return appendRow(s.filePath, s.toRow(p))
}

func (s *PositionStore) Update(p *models.Position) error {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return err
	}
	for i, row := range rows {
		rowID, _ := strconv.Atoi(row[0])
		if rowID == p.ID {
			rows[i] = s.toRow(p)
			return writeAllRows(s.filePath, positionHeader, rows)
		}
	}
	return fmt.Errorf("position not found")
}

func (s *PositionStore) Delete(id int) error {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return err
	}
	var newRows [][]string
	for _, row := range rows {
		rowID, _ := strconv.Atoi(row[0])
		if rowID != id {
			newRows = append(newRows, row)
		}
	}
	return writeAllRows(s.filePath, positionHeader, newRows)
}

func (s *PositionStore) DeleteByEventID(eventID int) error {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return err
	}
	var newRows [][]string
	for _, row := range rows {
		eid, _ := strconv.Atoi(row[2])
		if eid != eventID {
			newRows = append(newRows, row)
		}
	}
	return writeAllRows(s.filePath, positionHeader, newRows)
}
