package store

import (
	"encoding/csv"
	"os"
	"strconv"
)

func readAllRows(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	// Skip header
	if len(rows) > 0 {
		return rows[1:], nil
	}
	return nil, nil
}

func writeAllRows(filePath string, header []string, rows [][]string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write(header); err != nil {
		return err
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

func appendRow(filePath string, row []string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()
	return w.Write(row)
}

func nextID(filePath string) (int, error) {
	rows, err := readAllRows(filePath)
	if err != nil {
		return 1, nil
	}
	if len(rows) == 0 {
		return 1, nil
	}
	lastID, err := strconv.Atoi(rows[len(rows)-1][0])
	if err != nil {
		return 1, nil
	}
	return lastID + 1, nil
}
