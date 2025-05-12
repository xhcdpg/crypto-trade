package position

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/xhcdpg/crypto-trade/models"
	"github.com/xhcdpg/crypto-trade/types"
	u "github.com/xhcdpg/crypto-trade/user"
	"sync"
)

type PositionManager struct {
	positions map[string]map[string]*models.Position
	publisher message.Publisher
	mutex     sync.Mutex
}

func NewPositionManager(publisher message.Publisher) *PositionManager {
	return &PositionManager{
		positions: make(map[string]map[string]*models.Position),
		publisher: publisher,
	}
}

func (pm *PositionManager) getOrCreatePosition(userID, symbol string) *models.Position {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if _, ok := pm.positions[userID]; !ok {
		pm.positions[userID] = make(map[string]*models.Position)
	}

	if position, ok := pm.positions[userID][symbol]; ok {
		return position
	}

	newPosition := &models.Position{
		ID:                uuid.New().String(),
		UserID:            userID,
		Symbol:            symbol,
		Quantity:          0.0,
		EntryPrice:        0.0,
		MarkPrice:         0.0,
		UnrealizedPnl:     0.0,
		RealizedPnl:       0.0,
		AllocatedMargin:   0.0,
		Leverage:          1,
		MaintenanceMargin: 0.0,
		InitialMargin:     0.0,
		LiquidationPrice:  0.0,
	}
	pm.positions[userID][symbol] = newPosition
	return newPosition
}

func (pm *PositionManager) GetPosition(userID, symbol string) *models.Position {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	if userPositions, ok := pm.positions[userID]; ok {
		if position, ok := userPositions[symbol]; ok {
			return position
		}
	}
	return nil
}

func (pm *PositionManager) GetAllPositions() []*models.Position {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	var allPositions []*models.Position
	for _, userPositions := range pm.positions {
		for _, position := range userPositions {
			if position.Quantity != 0 {
				allPositions = append(allPositions, position)
			}
		}
	}

	return allPositions
}

func (pm *PositionManager) UpdatePositionFromTrade(trade *models.Trade, leverage uint, marginType types.MarginMode) error {
	var (
		position *models.Position
		side     types.Side
	)

	if trade.BuyerID != types.SystemID {
		position = pm.getOrCreatePosition(trade.BuyerID, trade.Symbol)
		side = types.Buy
	} else {
		position = pm.getOrCreatePosition(trade.SellerID, trade.Symbol)
		side = types.Sell
	}

	user, err := u.GlobalUserService.GetUser(trade.ID)
	if err != nil {
		return err
	}

	if position.Quantity == 0 {
		// new position
		position.EntryPrice = trade.Price
	} else {
		if position.Side == types.Buy {
			if side == types.Buy {
				totalQuantity := position.Quantity + trade.Quantity
				position.EntryPrice = (position.EntryPrice*position.Quantity + trade.Price*trade.Quantity) / totalQuantity
			} else {
				// reverse open position
				// close position
				if position.Quantity > trade.Quantity {
					position.Quantity -= trade.Quantity
					position.RealizedPnl += (trade.Price - position.EntryPrice) * trade.Quantity
				} else {

				}
			}
		} else {
			if side == types.Buy {

			} else {

			}
		}
	}

	return nil
}
