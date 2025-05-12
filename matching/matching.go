package matching

import (
	"container/heap"
	"encoding/json"
	"errors"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/xhcdpg/crypto-trade/models"
	"github.com/xhcdpg/crypto-trade/position"
	"github.com/xhcdpg/crypto-trade/types"
	u "github.com/xhcdpg/crypto-trade/user"
	"time"
)

type OrderNode struct {
	Price     float64
	Quantity  float64
	OrderID   string
	UserID    string
	Timestamp time.Time
}

type BidsQueue []*OrderNode

func (b BidsQueue) Len() int      { return len(b) }
func (b BidsQueue) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b BidsQueue) Less(i, j int) bool {
	if b[i].Price == b[j].Price {
		return b[i].Timestamp.Before(b[j].Timestamp)
	}
	return b[i].Price > b[j].Price
}

func (b *BidsQueue) Push(x interface{}) {
	*b = append(*b, x.(*OrderNode))
}

func (b *BidsQueue) Pop() interface{} {
	old := *b
	n := len(old)
	x := old[n-1]
	*b = old[0 : n-1]
	return x
}

type AsksQueue []*OrderNode

func (a AsksQueue) Len() int      { return len(a) }
func (a AsksQueue) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a AsksQueue) Less(i, j int) bool {
	if a[i].Price == a[j].Price {
		return a[i].Timestamp.Before(a[j].Timestamp)
	}
	return a[i].Price < a[j].Price
}
func (a *AsksQueue) Push(x interface{}) {
	*a = append(*a, x.(*OrderNode))
}

func (a *AsksQueue) Pop() interface{} {
	old := *a
	n := len(old)
	x := old[n-1]
	*a = old[0 : n-1]
	return x
}

type StopQueue struct {
	Orders []*models.Order
}

type OrderBook struct {
	Symbol string
	Bids   BidsQueue
	Asks   AsksQueue
	Stops  *StopQueue
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		Bids:   make(BidsQueue, 0),
		Asks:   make(AsksQueue, 0),
		Stops:  &StopQueue{Orders: []*models.Order{}},
	}
}

func (ob *OrderBook) GetMidPrice() float64 {
	if len(ob.Asks) == 0 || len(ob.Bids) == 0 {
		return 0.0
	}
	return (ob.Asks[0].Price + ob.Bids[0].Price) / 2
}

type MatchingEngine struct {
	orderBooks      map[string]*OrderBook
	positionManager *position.PositionManager
}

func (m *MatchingEngine) GetOrderBook(symbol string) *OrderBook {
	if ob, ok := m.orderBooks[symbol]; ok {
		return ob
	}
	ob := NewOrderBook(symbol)
	m.orderBooks[symbol] = ob
	return ob
}

func (m *MatchingEngine) GetCurrentPrice(symbol string) float64 {
	ob := m.GetOrderBook(symbol)
	return ob.GetMidPrice()
}

func (m *MatchingEngine) PlaceOrder(order *models.Order, publisher message.Publisher) error {
	// todo: validate order
	user, err := u.GlobalUserService.GetUser(order.UserID)
	if err != nil {
		return err
	}

	if user.MarginMode == types.IsolatedMargin && order.Type != types.Market && order.Type != types.Limit {
		return errors.New("only market and limit order are supported on isolated margin mode")
	}

	ob := m.GetOrderBook(order.Symbol)
	if user.MarginMode == types.CrossMargin {
		if err := m.checkCrossMargin(user, order); err != nil {
			return err
		}
	} else {
		if err := m.checkIsolatedMargin(user, order); err != nil {
			return err
		}
	}

	orderJson, err := json.Marshal(order)
	if err != nil {
		return err
	}

	err = publisher.Publish("orders", message.NewMessage(uuid.New().String(), orderJson))
	switch order.Type {
	case types.Limit:
		err = m.handleLimitOrder(ob, order, publisher)
	case types.Market:
		err = m.handleMarketOrder(ob, order, publisher)
	case types.LimitStopLoss, types.LimitTakeProfit, types.MarketStopLoss, types.MarketTakeProfit:
		ob.Stops.Orders = append(ob.Stops.Orders, order)
		order.Status = types.Pending
	}

	return err
}

func (m *MatchingEngine) checkCrossMargin(user *models.User, order *models.Order) error {
	entryPrice := m.GetCurrentPrice(order.Symbol)
	if entryPrice == 0.0 {
		return errors.New("cannot get current price")
	}

	margin := (order.Quantity * entryPrice) / float64(order.Leverage)
	totalMargin := margin
	for _, p := range user.Positions {
		pMargin := (p.Quantity * p.EntryPrice) / float64(p.Leverage)
		totalMargin += pMargin
	}

	if totalMargin > user.TotalBalance {
		return errors.New("insufficient balance to open position")
	}
	return nil
}

func (m *MatchingEngine) checkIsolatedMargin(user *models.User, order *models.Order) error {
	entryPrice := m.GetCurrentPrice(order.Symbol)
	if entryPrice == 0.0 {
		return errors.New("cannot get current price")
	}

	allocatedMargin := order.Quantity * entryPrice * 0.1
	currentSumAllocated := 0.0
	for _, p := range user.Positions {
		currentSumAllocated += p.AllocatedMargin
	}
	if user.TotalBalance < currentSumAllocated+allocatedMargin {
		return errors.New("insufficient balance to open position")
	}
	return nil
}

func (m *MatchingEngine) handleLimitOrder(ob *OrderBook, order *models.Order, publisher message.Publisher) error {
	var err error
	if order.Side == types.Buy {
		err = m.matchBuyLimit(ob, order, publisher)
	} else if order.Side == types.Sell {
		err = m.matchSellLimit(ob, order, publisher)
	}
	return err
}

func (m *MatchingEngine) matchBuyLimit(ob *OrderBook, order *models.Order, publisher message.Publisher) error {
	markPrice := m.GetCurrentPrice(order.Symbol)
	if markPrice == 0.0 {
		return errors.New("cannot get current price")
	}
	if order.Price < markPrice {
		node := &OrderNode{
			Price:     order.Price,
			Quantity:  order.Quantity,
			OrderID:   order.ID,
			UserID:    order.UserID,
			Timestamp: order.Timestamp,
		}
		order.Status = types.Pending
		heap.Push(&ob.Bids, node)
		return nil
	}

	trade := &models.Trade{
		ID:        uuid.New().String(),
		Symbol:    order.Symbol,
		BuyerID:   types.SystemID,
		SellerID:  order.ID,
		Price:     markPrice,
		Quantity:  order.Quantity,
		Timestamp: order.Timestamp,
	}

	tradeJson, err := json.Marshal(trade)
	if err != nil {
		return err
	}
	err = publisher.Publish("trades", message.NewMessage(uuid.New().String(), tradeJson))
	if err != nil {
		return err
	}
	order.Status = types.Filled
	heap.Pop(&ob.Bids)

	return nil
}

func (m *MatchingEngine) matchSellLimit(ob *OrderBook, order *models.Order, publisher message.Publisher) error {
	markPrice := m.GetCurrentPrice(order.Symbol)
	if markPrice == 0.0 {
		return errors.New("cannot get current price")
	}

	if order.Price > markPrice {
		node := &OrderNode{
			Price:     order.Price,
			Quantity:  order.Quantity,
			OrderID:   order.ID,
			UserID:    order.UserID,
			Timestamp: order.Timestamp,
		}
		order.Status = types.Open
		heap.Push(&ob.Asks, node)
		return nil
	}

	trade := &models.Trade{
		ID:        uuid.New().String(),
		Symbol:    order.Symbol,
		BuyerID:   types.SystemID,
		SellerID:  order.ID,
		Price:     order.Price,
		Quantity:  order.Quantity,
		Timestamp: time.Now(),
	}
	tradeJson, err := json.Marshal(trade)
	if err != nil {
		return err
	}
	err = publisher.Publish("trades", message.NewMessage(uuid.New().String(), tradeJson))
	if err != nil {
		return err
	}

	order.Status = types.Filled
	heap.Pop(&ob.Asks)

	return nil
}

func (m *MatchingEngine) handleMarketOrder(ob *OrderBook, order *models.Order, publisher message.Publisher) error {
	markPrice := m.GetCurrentPrice(order.Symbol)
	if markPrice == 0.0 {
		return errors.New("cannot get current price")
	}

	if order.Side == types.Buy {
		trade := &models.Trade{
			ID:        uuid.New().String(),
			Symbol:    order.Symbol,
			BuyerID:   order.UserID,
			SellerID:  types.SystemID,
			Price:     markPrice,
			Quantity:  order.Quantity,
			Timestamp: time.Now(),
		}
		tradJson, err := json.Marshal(trade)
		if err != nil {
			return err
		}

		err = publisher.Publish("trades", message.NewMessage(uuid.New().String(), tradJson))
		if err != nil {
			return err
		}

		m.positionManager.UpdatePositionFromTrade(trade, order.Leverage, order.MarginType)
		order.Quantity = 0.0
		order.Status = types.Filled
		heap.Pop(&ob.Bids)
	} else if order.Side == types.Sell {
		trade := &models.Trade{
			ID:        uuid.New().String(),
			Symbol:    order.Symbol,
			BuyerID:   types.SystemID,
			SellerID:  order.UserID,
			Price:     markPrice,
			Quantity:  order.Quantity,
			Timestamp: time.Now(),
		}
		tradJson, err := json.Marshal(trade)
		if err != nil {
			return err
		}

		err = publisher.Publish("trades", message.NewMessage(uuid.New().String(), tradJson))
		if err != nil {
			return err
		}

		m.positionManager.UpdatePositionFromTrade(trade, order.Leverage, order.MarginType)
		order.Quantity = 0.0
		order.Status = types.Filled
		heap.Pop(&ob.Asks)
	}

	return nil
}

func (m *MatchingEngine) MonitorStops(publisher message.Publisher) {
	for _, ob := range m.orderBooks {
		midPrice := ob.GetMidPrice()
		for i := 0; i < len(ob.Stops.Orders); i++ {
			order := ob.Stops.Orders[i]
			if shouldTriggerStop(order, midPrice) {
				if order.Type == types.MarketStopLoss || order.Type == types.MarketTakeProfit {
					marketOrder := &models.Order{
						ID:         order.ID,
						UserID:     order.UserID,
						Symbol:     order.Symbol,
						Side:       order.Side,
						Type:       types.Market,
						Leverage:   order.Leverage,
						Quantity:   order.Quantity,
						Price:      0.0,
						StopPrice:  0,
						Status:     "",
						MarginType: order.MarginType,
						Timestamp:  time.Time{},
					}
					m.PlaceOrder(marketOrder, publisher)
					ob.Stops.Orders = append(ob.Stops.Orders[:i], ob.Stops.Orders[i+1:]...)
					i--
				} else if order.Type == types.LimitStopLoss || order.Type == types.LimitTakeProfit {
					limitOrder := &models.Order{
						ID:         order.ID,
						UserID:     order.UserID,
						Symbol:     order.Symbol,
						Side:       order.Side,
						Type:       types.Limit,
						Leverage:   order.Leverage,
						Quantity:   order.Quantity,
						Price:      order.Price,
						StopPrice:  0,
						Status:     "",
						MarginType: order.MarginType,
						Timestamp:  time.Now(),
					}
					m.PlaceOrder(limitOrder, publisher)
					ob.Stops.Orders = append(ob.Stops.Orders[:i], ob.Stops.Orders[i+1:]...)
					i--
				}
			}
		}
	}
}

func shouldTriggerStop(order *models.Order, currentPrice float64) bool {
	if order.Type == types.LimitStopLoss || order.Type == types.MarketStopLoss {
		if order.Side == types.Buy && currentPrice >= order.StopPrice {
			return true
		}
		if order.Side == types.Sell && currentPrice <= order.StopPrice {
			return true
		}
	}
	if order.Type == types.LimitTakeProfit || order.Type == types.MarketTakeProfit {
		if order.Side == types.Buy && currentPrice <= order.StopPrice {
			return true
		}
		if order.Side == types.Sell && currentPrice >= order.StopPrice {
			return true
		}
	}
	return false
}
