package store

import (
	"fmt"
	"pauls-bach/models"
	"strconv"
	"time"
)

type UserStore struct {
	filePath string
}

var userHeader = []string{"id", "username", "pin_hash", "balance", "is_admin", "bingo", "created_at"}

func (s *UserStore) toRow(u *models.User) []string {
	return []string{
		strconv.Itoa(u.ID),
		u.Username,
		u.PinHash,
		strconv.Itoa(u.Balance),
		strconv.FormatBool(u.IsAdmin),
		strconv.FormatBool(u.Bingo),
		u.CreatedAt,
	}
}

func (s *UserStore) fromRow(row []string) (*models.User, error) {
	id, _ := strconv.Atoi(row[0])
	balance, _ := strconv.Atoi(row[3])
	isAdmin := row[4] == "true"
	bingo := false
	createdAt := row[5]
	if len(row) > 6 {
		bingo = row[5] == "true"
		createdAt = row[6]
	}
	return &models.User{
		ID:        id,
		Username:  row[1],
		PinHash:   row[2],
		Balance:   balance,
		IsAdmin:   isAdmin,
		Bingo:     bingo,
		CreatedAt: createdAt,
	}, nil
}

func (s *UserStore) GetAll() ([]models.User, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	users := make([]models.User, 0, len(rows))
	for _, row := range rows {
		u, err := s.fromRow(row)
		if err != nil {
			continue
		}
		users = append(users, *u)
	}
	return users, nil
}

func (s *UserStore) GetByID(id int) (*models.User, error) {
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
	return nil, fmt.Errorf("user not found")
}

func (s *UserStore) GetByUsername(username string) (*models.User, error) {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row[1] == username {
			return s.fromRow(row)
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (s *UserStore) Create(u *models.User) error {
	id, err := nextID(s.filePath)
	if err != nil {
		return err
	}
	u.ID = id
	u.CreatedAt = time.Now().Format(time.RFC3339)
	return appendRow(s.filePath, s.toRow(u))
}

func (s *UserStore) Update(u *models.User) error {
	rows, err := readAllRows(s.filePath)
	if err != nil {
		return err
	}
	for i, row := range rows {
		rowID, _ := strconv.Atoi(row[0])
		if rowID == u.ID {
			rows[i] = s.toRow(u)
			return writeAllRows(s.filePath, userHeader, rows)
		}
	}
	return fmt.Errorf("user not found")
}
