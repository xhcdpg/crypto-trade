package position

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/xhcdpg/crypto-trade/models"
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
