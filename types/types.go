package types

type MarginMode string

const (
	CrossMargin    MarginMode = "cross"
	IsolatedMargin MarginMode = "isolated"
)

type OrderType string

const (
	Limit            OrderType = "limit"
	Market           OrderType = "market"
	LimitStopLoss    OrderType = "limit_stop_loss"
	LimitTakeProfit  OrderType = "limit_take_profit"
	MarketStopLoss   OrderType = "market_stop_loss"
	MarketTakeProfit OrderType = "market_take_profit"
)

type OrderStatus string

const (
	Open      OrderStatus = "open" // 待撮合
	Filled    OrderStatus = "filled"
	Cancelled OrderStatus = "cancelled"
	Pending   OrderStatus = "pending" // 待激活
)

type Side string

const (
	Sell Side = "sell"
	Buy  Side = "buy"
)
