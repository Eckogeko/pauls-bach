package store

import (
	"pauls-bach/models"
	"strconv"
	"time"
)

type TransactionStore struct {
	filePath string
}

func (s *TransactionStore) toRow(t *models.Transaction) []string {
	return []string{
		strconv.Itoa(t.ID),
		strconv.Itoa(t.UserID),
		strconv.Itoa(t.EventID),
		strconv.Itoa(t.OutcomeID),
		t.TxType,
		strconv.FormatFloat(t.Shares, 'f', 6, 64),
		strconv.Itoa(t.Points),
		t.CreatedAt,
	}
}

func (s *TransactionStore) fromRow(row []string) *models.Transaction {
	id, _ := strconv.Atoi(row[0])
	userID, _ := strconv.Atoi(row[1])
	eventID, _ := strconv.Atoi(row[2])
	outcomeID, _ := strconv.Atoi(row[3])
	shares, _ := strconv.ParseFloat(row[5], 64)
	points, _ := strconv.Atoi(row[6])
	return &models.Transaction{
		ID:        id,
		UserID:    userID,
		EventID:   eventID,
		OutcomeID: outcomeID,
		TxType:    row[4],
		Shares:    shares,
		Points:    points,
		CreatedAt: row[7],
	}
}

func (s *TransactionStore) GetByUserID(userID int) ([]models.Transaction, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	var txs []models.Transaction
	for _, row := range rows {
		uid, _ := strconv.Atoi(row[1])
		if uid == userID {
			txs = append(txs, *s.fromRow(row))
		}
	}
	return txs, nil
}

func (s *TransactionStore) Create(t *models.Transaction) error {
	id, err := nextID(s.filePath)
	if err != nil {
		return err
	}
	t.ID = id
	t.CreatedAt = time.Now().Format(time.RFC3339)
	return appendRow(s.filePath, s.toRow(t))
}
