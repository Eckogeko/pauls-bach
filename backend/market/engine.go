package market

import (
	"fmt"
	"math"
	"time"

	"pauls-bach/models"
	"pauls-bach/store"
)

type Engine struct {
	Store *store.Store
}

type UserOutcome struct {
	UserID int  `json:"user_id"`
	Won    bool `json:"won"`
	Payout int  `json:"payout"`
	Refund bool `json:"refund"`
}

type ResolveResult struct {
	UserOutcomes []UserOutcome
}

type OutcomeOdds struct {
	OutcomeID int     `json:"outcome_id"`
	Label     string  `json:"label"`
	Odds      float64 `json:"odds"` // 0-100 percentage
	Shares    float64 `json:"shares"`
}

func (e *Engine) GetOdds(eventID int) ([]OutcomeOdds, error) {
	outcomes, err := e.Store.Outcomes.GetByEventID(eventID)
	if err != nil {
		return nil, err
	}
	positions, err := e.Store.Positions.GetByEventID(eventID)
	if err != nil {
		return nil, err
	}

	sharesByOutcome := make(map[int]float64)
	for _, p := range positions {
		sharesByOutcome[p.OutcomeID] += p.Shares
	}

	var totalShares float64
	for _, s := range sharesByOutcome {
		totalShares += s
	}

	odds := make([]OutcomeOdds, len(outcomes))
	for i, o := range outcomes {
		s := sharesByOutcome[o.ID]
		var pct float64
		if totalShares == 0 {
			pct = 100.0 / float64(len(outcomes))
		} else {
			pct = (s / totalShares) * 100
		}
		odds[i] = OutcomeOdds{
			OutcomeID: o.ID,
			Label:     o.Label,
			Odds:      math.Round(pct*100) / 100,
			Shares:    s,
		}
	}
	return odds, nil
}

func (e *Engine) Buy(userID, eventID, outcomeID, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	user, err := e.Store.Users.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if user.Balance < amount {
		return fmt.Errorf("insufficient balance")
	}

	event, err := e.Store.Events.GetByID(eventID)
	if err != nil {
		return fmt.Errorf("event not found")
	}
	if event.Status != "open" {
		return fmt.Errorf("event is not open for betting")
	}

	// Verify outcome belongs to event
	outcomes, err := e.Store.Outcomes.GetByEventID(eventID)
	if err != nil {
		return err
	}
	validOutcome := false
	for _, o := range outcomes {
		if o.ID == outcomeID {
			validOutcome = true
			break
		}
	}
	if !validOutcome {
		return fmt.Errorf("invalid outcome for this event")
	}

	// Check if user already has a position on a different outcome for this event
	existingPositions, err := e.Store.Positions.GetByUserAndEvent(userID, eventID)
	if err != nil {
		return err
	}
	for _, p := range existingPositions {
		if p.OutcomeID != outcomeID {
			return fmt.Errorf("you already bet on a different outcome for this event")
		}
	}

	// Calculate current price for avg_price tracking
	odds, err := e.GetOdds(eventID)
	if err != nil {
		return err
	}
	var currentPrice float64
	for _, o := range odds {
		if o.OutcomeID == outcomeID {
			currentPrice = o.Odds / 100
			break
		}
	}
	if currentPrice == 0 {
		currentPrice = 1.0 / float64(len(outcomes))
	}

	shares := float64(amount) // 1 point = 1 share

	// Update or create position
	pos, err := e.Store.Positions.GetByUserEventOutcome(userID, eventID, outcomeID)
	if err != nil {
		return err
	}
	if pos != nil {
		// Update existing: recalculate avg price
		totalCost := pos.AvgPrice*pos.Shares + currentPrice*shares
		pos.Shares += shares
		pos.AvgPrice = totalCost / pos.Shares
		if err := e.Store.Positions.Update(pos); err != nil {
			return err
		}
	} else {
		pos = &models.Position{
			UserID:    userID,
			EventID:   eventID,
			OutcomeID: outcomeID,
			Shares:    shares,
			AvgPrice:  currentPrice,
		}
		if err := e.Store.Positions.Create(pos); err != nil {
			return err
		}
	}

	// Deduct balance
	user.Balance -= amount
	if err := e.Store.Users.Update(user); err != nil {
		return err
	}

	// Record transaction
	tx := &models.Transaction{
		UserID:    userID,
		EventID:   eventID,
		OutcomeID: outcomeID,
		TxType:    "buy",
		Shares:    shares,
		Points:    amount,
	}
	return e.Store.Transactions.Create(tx)
}

func (e *Engine) Sell(userID, eventID, outcomeID int, sharesToSell float64) (int, error) {
	if sharesToSell <= 0 {
		return 0, fmt.Errorf("shares must be positive")
	}
	if sharesToSell != math.Floor(sharesToSell) {
		return 0, fmt.Errorf("shares must be a whole number")
	}

	event, err := e.Store.Events.GetByID(eventID)
	if err != nil {
		return 0, fmt.Errorf("event not found")
	}
	if event.Status != "open" {
		return 0, fmt.Errorf("event is not open for trading")
	}

	pos, err := e.Store.Positions.GetByUserEventOutcome(userID, eventID, outcomeID)
	if err != nil {
		return 0, err
	}
	if pos == nil || pos.Shares < sharesToSell {
		return 0, fmt.Errorf("insufficient shares")
	}

	// Seller gets 50% of share value back; the rest stays in the prize pool
	fullValue := int(math.Floor(sharesToSell))
	pointsBack := fullValue / 2
	if pointsBack < 1 && sharesToSell > 0 {
		pointsBack = 1
	}

	// Remove shares from user's position but keep them in the pool
	// by not deleting from total shares — leave orphan shares so pool stays large
	pos.Shares -= sharesToSell
	if pos.Shares < 0.001 {
		if err := e.Store.Positions.Delete(pos.ID); err != nil {
			return 0, err
		}
	} else {
		if err := e.Store.Positions.Update(pos); err != nil {
			return 0, err
		}
	}

	// Credit user with 50%
	user, err := e.Store.Users.GetByID(userID)
	if err != nil {
		return 0, err
	}
	user.Balance += pointsBack
	if err := e.Store.Users.Update(user); err != nil {
		return 0, err
	}

	// Record transaction
	tx := &models.Transaction{
		UserID:    userID,
		EventID:   eventID,
		OutcomeID: outcomeID,
		TxType:    "sell",
		Shares:    sharesToSell,
		Points:    pointsBack,
	}
	if err := e.Store.Transactions.Create(tx); err != nil {
		return 0, err
	}

	return pointsBack, nil
}

func (e *Engine) Resolve(eventID, winningOutcomeID int) (*ResolveResult, error) {
	event, err := e.Store.Events.GetByID(eventID)
	if err != nil {
		return nil, fmt.Errorf("event not found")
	}
	if event.Status == "resolved" {
		return nil, fmt.Errorf("event already resolved")
	}

	positions, err := e.Store.Positions.GetByEventID(eventID)
	if err != nil {
		return nil, err
	}

	result := &ResolveResult{}

	// Calculate total pool and winning shares
	var totalPool float64
	var winningShares float64
	for _, p := range positions {
		totalPool += p.Shares
		if p.OutcomeID == winningOutcomeID {
			winningShares += p.Shares
		}
	}

	// Track which users we've already recorded (a user may have multiple positions)
	seen := make(map[int]bool)

	if totalPool > 0 {
		if winningShares == 0 {
			// No one bet on the winner - refund everyone proportionally
			for _, p := range positions {
				user, err := e.Store.Users.GetByID(p.UserID)
				if err != nil {
					continue
				}
				refund := int(math.Round(p.Shares))
				user.Balance += refund
				e.Store.Users.Update(user)
				e.Store.Transactions.Create(&models.Transaction{
					UserID:    p.UserID,
					EventID:   eventID,
					OutcomeID: p.OutcomeID,
					TxType:    "payout",
					Shares:    p.Shares,
					Points:    refund,
				})
				if !seen[p.UserID] {
					seen[p.UserID] = true
					result.UserOutcomes = append(result.UserOutcomes, UserOutcome{
						UserID: p.UserID,
						Won:    false,
						Payout: refund,
						Refund: true,
					})
				}
			}
		} else {
			// Record losers
			for _, p := range positions {
				if p.OutcomeID == winningOutcomeID {
					continue
				}
				if !seen[p.UserID] {
					seen[p.UserID] = true
					result.UserOutcomes = append(result.UserOutcomes, UserOutcome{
						UserID: p.UserID,
						Won:    false,
						Payout: 0,
						Refund: false,
					})
				}
			}

			// Distribute pool to winners proportionally + 50pt bonus
			for _, p := range positions {
				if p.OutcomeID != winningOutcomeID {
					continue
				}
				user, err := e.Store.Users.GetByID(p.UserID)
				if err != nil {
					continue
				}
				payout := int(math.Round(totalPool * (p.Shares / winningShares)))
				user.Balance += payout + 50
				e.Store.Users.Update(user)
				e.Store.Transactions.Create(&models.Transaction{
					UserID:    p.UserID,
					EventID:   eventID,
					OutcomeID: p.OutcomeID,
					TxType:    "payout",
					Shares:    p.Shares,
					Points:    payout,
				})
				e.Store.Transactions.Create(&models.Transaction{
					UserID:    p.UserID,
					EventID:   eventID,
					OutcomeID: p.OutcomeID,
					TxType:    "bonus",
					Shares:    0,
					Points:    50,
				})
				if !seen[p.UserID] {
					seen[p.UserID] = true
					result.UserOutcomes = append(result.UserOutcomes, UserOutcome{
						UserID: p.UserID,
						Won:    true,
						Payout: payout + 50,
						Refund: false,
					})
				}
			}
		}
	}

	// Clean up positions
	e.Store.Positions.DeleteByEventID(eventID)

	// Update event status
	event.Status = "resolved"
	event.WinningOutcomeID = winningOutcomeID
	event.ResolvedAt = time.Now().Format(time.RFC3339)
	return result, e.Store.Events.Update(event)
}
