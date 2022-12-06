package okex

import (
	"github.com/xiaolo66/He_Quan"
	"strconv"
	"strings"
	"time"
	."github.com/xiaolo66/He_Quan/utils"

)

type Stream struct {
	Op   string   `json:"op"`
	Args []string `json:"args"`
}

func SubscribeStream(event ...string) Stream {
	return Stream{
		Op:   "subscribe",
		Args: append([]string{}, event...),
	}
}

func UnSubscribeStream(event ...string) Stream {
	return Stream{
		Op:   "unsubscribe",
		Args: append([]string{}, event...),
	}
}

func LoginStream(event ...string) Stream {
	return Stream{
		Op:   "login",
		Args: append([]string{}, event...),
	}
}

// ResponseEvent
type ResponseEvent struct {
	Event     string `json:"event"`
	Channel   string `json:"channel"`
	Table     string `json:"table"`
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}

// OrderBook
type OrderBook struct {
	He_Quan.OrderBook
	Timestamp time.Time `json:"lastTime"` //
}

func (o *OrderBook) update(bookData DepthData) {
	o.Bids = o.Bids.Update(bookData.Bids, true)
	o.Asks = o.Asks.Update(bookData.Asks, false)
}

// OrderBook of one symbol
type SymbolOrderBook map[string]*OrderBook

type DepthData struct {
	Checksum  int32         `json:"checksum"`
	Symbol    string        `json:"instrument_id"`
	Timestamp time.Time     `json:"timestamp"`
	Bids      He_Quan.RawDepth `json:"bids"`
	Asks      He_Quan.RawDepth `json:"asks"`
}

type OrderBookRes struct {
	Action string      `json:"action"`
	Data   []DepthData `json:"data"`
}

type Ticker struct {
	Symbol      string `json:"instrument_id"`
	BestBid     string `json:"best_bid"`
	BestBidSize string `json:"best_bid_size"`
	BestAsk     string `json:"best_ask"`
	BestAskSize string `json:"best_ask_size"`
	High        string `json:"high_24h"`
	Last        string `json:"last"`
	Low         string `json:"low_24h"`
	Open        string `json:"open_24h"`
	Vol         string `json:"base_volume_24h"`
	Timestamp   string `json:"timestamp"`
}

func (t Ticker) parseTicker(symbol string) He_Quan.Ticker {
	return He_Quan.Ticker{
		Symbol:        symbol,
		Timestamp:     ParseIsoTime(t.Timestamp, nil),
		BestBuyPrice:  SafeParseFloat(t.BestBid),
		BestSellPrice: SafeParseFloat(t.BestAsk),
		Open:          SafeParseFloat(t.Open),
		Last:          SafeParseFloat(t.Last),
		High:          SafeParseFloat(t.High),
		Low:           SafeParseFloat(t.Low),
		Vol:           SafeParseFloat(t.Vol),
	}
}

type TickerRes struct {
	Data []Ticker `json:"data"`
}

type Trade struct {
	Symbol    string `json:"instrument_id"`
	Timestamp string `json:"timestamp"`
	Price     string `json:"price"`
	Size      string `json:"size"`
	Side      string `json:"side"`
}

func (t Trade) parseTrade(symbol string) He_Quan.Trade {
	var side He_Quan.Side = He_Quan.Buy
	if t.Side == "sell" {
		side = He_Quan.Sell
	}
	return He_Quan.Trade{
		Symbol:    symbol,
		Timestamp: ParseIsoTime(t.Timestamp, nil),
		Price:     SafeParseFloat(t.Price),
		Amount:    SafeParseFloat(t.Size),
		Side:      side,
	}
}

type TradeRes struct {
	Data []Trade `json:"data"`
}

type KLine [6]string

func (k KLine) parseKLine(symbol string) He_Quan.KLine {
	return He_Quan.KLine{
		Symbol:    symbol,
		Timestamp: ParseIsoTime(k[0], nil),
		Open:      SafeParseFloat(k[1]),
		High:      SafeParseFloat(k[2]),
		Close:     SafeParseFloat(k[3]),
		Low:       SafeParseFloat(k[4]),
		Volume:    SafeParseFloat(k[5]),
	}
}

type KLineRes struct {
	Data []struct {
		Candle KLine  `json:"candle"`
		Symbol string `json:"instrument_id"`
	} `json:"data"`
}

type Balance struct {
	Balance   string `json:"balance"`
	Available string `json:"available"`
	Currency  string `json:"currency"`
	Hold      string `json:"hold"`
	Timestamp string `json:"timestamp"`
}

func (b Balance) parseBalance() He_Quan.Balance {
	return He_Quan.Balance{
		Asset:     strings.ToUpper(b.Currency),
		Available: SafeParseFloat(b.Available),
		Frozen:    SafeParseFloat(b.Hold),
	}
}

type BalanceRes struct {
	Data []Balance `json:"data"`
}

type Order struct {
	Symbol         string `json:"instrument_id"`
	OrderId        string `json:"order_id"`
	ClientOId      string `json:"client_oid"`
	Price          string `json:"price"`
	Size           string `json:"size"`
	Side           string `json:"side"` //Buy or sell
	Type           string `json:"type"` //limit,market(defaulted as limit)
	FilledSize     string `json:"filled_size"`
	FilledNotional string `json:"filled_notional"`
	OrderType      string `json:"order_type"` //0: Normal limit order 1: Post only 2: Fill Or Kill 3: Immediatel Or Cancel
	State          string `json:"state"`      //-2:Failed,-1:Canceled,0:Open ,1:Partially Filled, 2:Fully Filled,3:Submitting,4:Cancelling
	Timestamp      string `json:"timestamp"`
	CreatedAt      string `json:"created_at"`
}

func (o Order) parseOrder(symbol string) He_Quan.Order {
	order := He_Quan.Order{
		ID:         o.OrderId,
		ClientID:   o.ClientOId,
		Symbol:     symbol,
		Price:      o.Price,
		Amount:     o.Size,
		Filled:     o.FilledSize,
		Cost:       o.FilledNotional,
		CreateTime: ParseIsoTime(o.CreatedAt, nil),
	}
	switch o.Side {
	case "sell":
		order.Side = He_Quan.Sell
	case "buy":
		order.Side = He_Quan.Buy
	}
	switch o.Type {
	case "limit":
		order.Type = He_Quan.LIMIT
	case "market":
		order.Type = He_Quan.MARKET
	}
	switch o.OrderType {
	case "0":
		order.OrderType = He_Quan.Normal
	case "1":
		order.OrderType = He_Quan.PostOnly
	case "2":
		order.OrderType = He_Quan.FOK
	case "3":
		order.OrderType = He_Quan.IOC
	}
	switch o.State {
	case "-1":
		order.Status = He_Quan.Canceled
		filled, err := strconv.ParseFloat(order.Filled, 64)
		if err == nil && filled > 0 {
			order.Status = He_Quan.Close
		}
	case "2":
		order.Status = He_Quan.Close
	case "0", "3":
		order.Status = He_Quan.Open
	case "1":
		order.Status = He_Quan.Partial
	default:
		order.Status = He_Quan.OrderStatusUnKnown
	}
	return order
}

type OrderRes struct {
	Data []Order `json:"data"`
}

