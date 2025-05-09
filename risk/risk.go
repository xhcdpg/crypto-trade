package risk

import (
	"github.com/xhcdpg/crypto-trade/models"
	"github.com/xhcdpg/crypto-trade/position"
)

type RiskManager struct {
	positionManager *position.PositionManager
}

func NewRiskManager(positionManager *position.PositionManager) *RiskManager {
	return &RiskManager{
		positionManager: positionManager,
	}
}

func (rm *RiskManager) GetAllPositions() []*models.Position {
	return rm.positionManager.GetAllPositions()
}
