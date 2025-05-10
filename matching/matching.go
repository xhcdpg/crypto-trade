package matching

import (
	"github.com/xhcdpg/crypto-trade/models"
	"github.com/xhcdpg/crypto-trade/position"
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
