package store

import (
	"pauls-bach/models"
	"strconv"
	"time"
)

type OddsSnapshotStore struct {
	filePath string
}

var oddsSnapshotHeader = []string{"id", "event_id", "outcome_id", "odds", "created_at"}

func (s *OddsSnapshotStore) toRow(o *models.OddsSnapshot) []string {
	return []string{
		strconv.Itoa(o.ID),
		strconv.Itoa(o.EventID),
		strconv.Itoa(o.OutcomeID),
		strconv.FormatFloat(o.Odds, 'f', 2, 64),
		o.CreatedAt,
	}
}

func (s *OddsSnapshotStore) fromRow(row []string) *models.OddsSnapshot {
	id, _ := strconv.Atoi(row[0])
	eventID, _ := strconv.Atoi(row[1])
	outcomeID, _ := strconv.Atoi(row[2])
	odds, _ := strconv.ParseFloat(row[3], 64)
	return &models.OddsSnapshot{
		ID:        id,
		EventID:   eventID,
		OutcomeID: outcomeID,
		Odds:      odds,
		CreatedAt: row[4],
	}
}

func (s *OddsSnapshotStore) DeleteByEventID(eventID int) error {
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
	return writeAllRows(s.filePath, oddsSnapshotHeader, kept)
}

func (s *OddsSnapshotStore) Create(o *models.OddsSnapshot) error {
	id, _ := nextID(s.filePath)
	o.ID = id
	if o.CreatedAt == "" {
		o.CreatedAt = time.Now().Format(time.RFC3339)
	}
	return appendRow(s.filePath, s.toRow(o))
}

func (s *OddsSnapshotStore) LastSnapshotTimeByEvent() (map[int]string, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	result := make(map[int]string)
	for _, row := range rows {
		o := s.fromRow(row)
		if existing, ok := result[o.EventID]; !ok || o.CreatedAt > existing {
			result[o.EventID] = o.CreatedAt
		}
	}
	return result, nil
}

func (s *OddsSnapshotStore) GetByEventID(eventID int) ([]*models.OddsSnapshot, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	var results []*models.OddsSnapshot
	for _, row := range rows {
		o := s.fromRow(row)
		if o.EventID == eventID {
			results = append(results, o)
		}
	}
	return results, nil
}
