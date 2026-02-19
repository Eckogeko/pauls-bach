package store

import (
	"os"
	"path/filepath"
	"sync"
)

var mu sync.RWMutex

func ReadLock()    { mu.RLock() }
func ReadUnlock()  { mu.RUnlock() }
func WriteLock()   { mu.Lock() }
func WriteUnlock() { mu.Unlock() }

type Store struct {
	Users         *UserStore
	Events        *EventStore
	Outcomes      *OutcomeStore
	Positions     *PositionStore
	Transactions  *TransactionStore
	OddsSnapshots *OddsSnapshotStore
	BingoEvents   *BingoEventStore
	BingoBoards   *BingoBoardStore
	BingoWinners  *BingoWinnerStore
	Activity      *ActivityStore
}

func New(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	headers := map[string]string{
		"users.csv":          "id,username,pin_hash,balance,is_admin,bingo,created_at",
		"events.csv":         "id,title,description,event_type,status,winning_outcome_id,created_at,resolved_at",
		"outcomes.csv":       "id,event_id,label",
		"positions.csv":      "id,user_id,event_id,outcome_id,shares,avg_price,created_at",
		"transactions.csv":   "id,user_id,event_id,outcome_id,tx_type,shares,points,created_at",
		"odds_snapshots.csv": "id,event_id,outcome_id,odds,created_at",
		"bingo_events.csv":   "id,title,resolved,created_at",
		"bingo_boards.csv":   "id,user_id,squares,created_at",
		"bingo_winners.csv":  "id,user_id,username,board_id,line,created_at",
		"activity.csv":       "id,type,message,user_id,event_id,created_at",
	}

	for file, header := range headers {
		p := filepath.Join(dataDir, file)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			if err := os.WriteFile(p, []byte(header+"\n"), 0644); err != nil {
				return nil, err
			}
		}
	}

	return &Store{
		Users:         &UserStore{filePath: filepath.Join(dataDir, "users.csv")},
		Events:        &EventStore{filePath: filepath.Join(dataDir, "events.csv")},
		Outcomes:      &OutcomeStore{filePath: filepath.Join(dataDir, "outcomes.csv")},
		Positions:     &PositionStore{filePath: filepath.Join(dataDir, "positions.csv")},
		Transactions:  &TransactionStore{filePath: filepath.Join(dataDir, "transactions.csv")},
		OddsSnapshots: &OddsSnapshotStore{filePath: filepath.Join(dataDir, "odds_snapshots.csv")},
		BingoEvents:   &BingoEventStore{filePath: filepath.Join(dataDir, "bingo_events.csv")},
		BingoBoards:   &BingoBoardStore{filePath: filepath.Join(dataDir, "bingo_boards.csv")},
		BingoWinners:  &BingoWinnerStore{filePath: filepath.Join(dataDir, "bingo_winners.csv")},
		Activity:      &ActivityStore{filePath: filepath.Join(dataDir, "activity.csv")},
	}, nil
}
