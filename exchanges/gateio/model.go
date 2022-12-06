package gateio

import (
	"He_Quan"
	"He_Quan/utils"
	"fmt"
	"time"

)

type Market struct {
	Symbol          string  `json:"id"`
	Base            string  `json:"base"`
	Quote           string  `json:"quote"`
	MinAmount       float64 `json:"min_quote_amount,string"`
	AmountPrecision int     `json:"amount_precision"`
	PricePrecision  int     `json:"precision"`
}

type ExchangeInfo []Market

type RawOrderBook struct {
	ID            int64         `json:"id,omitempty"`
	LastUpdateID  int64         `json:"last_update_id,omitempty" rest:"u"`
	FirstUpdateID int64         `rest:"U"`
	Asks          He_Quan.RawDepth `json:"asks" rest:"a"`
	Bids          He_Quan.RawDepth `json:"bids" rest:"b"`
	Symbol        string        `json:"s" rest:"s"`
}

type OrderBook struct {
	LastUpdateID int64 `json:"lastUpdateId"` // Last update ID
	He_Quan.OrderBook
}

func (r *RawOrderBook) parseOrderBook(symbol string) (orderBook He_Quan.OrderBook) {
	orderBook.Bids = orderBook.Bids.Update(r.Bids, true)
	orderBook.Asks = orderBook.Asks.Update(r.Asks, false)
	orderBook.Symbol = symbol
	return
}

type SymbolOrderBook map[string]*OrderBook

type Ticker struct {
	Timestamp  int64  `json:"etf_pre_timestamp"`
	Symbol     string `json:"currency_pair"`
	Last       string `json:"last"`
	High       string `json:"high_24h"`
	Percentage string `json:"change_percentage"`
	Low        string `json:"low_24h"`
	Vol        string `json:"base_volume"`
	BestBid    string `json:"highest_bid"`
	BestAsk    string `json:"lowest_ask"`
}

func (t *Ticker) parseTicker(symbol string) He_Quan.Ticker {
	tick := He_Quan.Ticker{
		Symbol:        symbol,
		Timestamp:     time.Duration(t.Timestamp),
		Last:          utils.SafeParseFloat(t.Last),
		Open:          utils.SafeParseFloat(t.Last) * (1 - utils.SafeParseFloat(t.Percentage)/100),
		BestBuyPrice:  utils.SafeParseFloat(t.BestBid),
		BestSellPrice: utils.SafeParseFloat(t.BestAsk),
		High:          utils.SafeParseFloat(t.High),
		Low:           utils.SafeParseFloat(t.Low),
		Vol:           utils.SafeParseFloat(t.Vol),
	}
	return tick
}

type Trade struct {
	Timestamp string `json:"create_time_ms"`
	Symbol    string `json:"currency_pair"`
	Side      string `json:"side"`
	Amount    string `json:"amount"`
	Price     string `json:"price"`
}

func (t *Trade) parseTrade(symbol string) He_Quan.Trade {
	trade := He_Quan.Trade{
		Symbol:    symbol,
		Timestamp: time.Duration(utils.SafeParseFloat(t.Timestamp)),
		Price:     utils.SafeParseFloat(t.Price),
		Amount:    utils.SafeParseFloat(t.Amount),
		Side:      He_Quan.Side(t.Side),
	}
	return trade
}

func parseKLienType(t He_Quan.KLineType) string {
	kt := ""
	switch t {
	case He_Quan.KLine1Minute:
		kt = "1m"
	case He_Quan.KLine5Minute:
		kt = "5m"
	case He_Quan.KLine15Minute:
		kt = "15m"
	case He_Quan.KLine30Minute:
		kt = "30m"
	case He_Quan.KLine1Hour:
		kt = "1h"
	case He_Quan.KLine4Hour:
		kt = "4h"
	case He_Quan.KLine8Hour:
		kt = "8h"
	case He_Quan.KLine1Day:
		kt = "1d"
	case He_Quan.KLine3Day:
		kt = "3d"
	case He_Quan.KLine1Week:
		kt = "7d"
	}
	return kt
}

type WsKline struct {
	Timestamp string `json:"t"`
	Volume    string `json:"v"`
	Close     string `json:"c"`
	High      string `json:"h"`
	Low       string `json:"l"`
	Open      string `json:"o"`
	IntSymbol string `json:"n"`
}

type Balance struct {
	Currency  string `json:"currency"`
	Available string `json:"available"`
	Locked    string `json:"locked"`
}

type WsBalance struct {
	Currency  string `json:"currency"`
	Total     string `json:"total"`
	Available string `json:"available"`
}

func (b *Balance) parserBalance() He_Quan.Balance {
	return He_Quan.Balance{
		Asset:     b.Currency,
		Available: utils.SafeParseFloat(b.Available),
		Frozen:    utils.SafeParseFloat(b.Locked),
	}
}

type Order struct {
	Id              string `json:"id" rest:"id"`
	Text            string `json:"text" rest:"text"`
	Symbol          string `json:"currency_pair" rest:"currency_pair"`
	CreateTime      string `json:"create_time" rest:"create_time"`
	TransactionTime string `json:"update_time" rest:"update_time"`
	Price           string `json:"price" rest:"price"`
	Amount          string `json:"amount" rest:"amount"`
	Left            string `json:"left" rest:"left"`
	Status          string `json:"status" rest:"event"`
	Side            string `json:"side" rest:"side"`
	Type            string `json:"type" rest:"type"`
	OrderType       string `json:"time_in_force" rest:"time_in_force"`
	Cost            string `json:"filled_total" rest:"filled_total"`
}

func (o *Order) parserOrder(symbol string) He_Quan.Order {
	order := He_Quan.Order{
		ID:              o.Id,
		ClientID:        o.Text,
		Symbol:          symbol,
		Price:           o.Price,
		Amount:          o.Amount,
		Filled:          fmt.Sprintf("%f", utils.SafeParseFloat(o.Amount)-utils.SafeParseFloat(o.Left)),
		Cost:            o.Cost,
		CreateTime:      time.Duration(utils.SafeParseFloat(o.CreateTime)),
		TransactionTime: time.Duration(utils.SafeParseFloat(o.TransactionTime)),
	}
	switch o.Side {
	case "buy":
		order.Side = He_Quan.Buy
	case "sell":
		order.Side = He_Quan.Sell
	}
	switch o.Type {
	case "limit":
		order.Type = He_Quan.LIMIT
	case "market":
		order.Type = He_Quan.MARKET
	}
	switch o.OrderType {
	case "ioc":
		order.OrderType = He_Quan.IOC
	case "poc":
		order.OrderType = He_Quan.PostOnly
	default:
		order.OrderType = He_Quan.Normal
	}
	switch o.Status {
	case "open", "put", "update":
		if utils.SafeParseFloat(order.Filled) > utils.ZERO {
			order.Status = He_Quan.Partial
		} else {
			order.Status = He_Quan.Open
		}
	case "cancelled", "finish":
		if utils.SafeParseFloat(order.Filled) > utils.ZERO {
			order.Status = He_Quan.Close
		} else {
			order.Status = He_Quan.Canceled
		}
	case "closed":
		order.Status = He_Quan.Close
	default:
		order.Status = He_Quan.OrderStatusUnKnown
	}
	return order
}

type ResponseEvent struct {
	Time    int64       `json:"time" rest:"time"`
	Channel string      `json:"channel" rest:"channel"`
	Event   string      `json:"event" rest:"event"`
	Error   interface{} `json:"error" rest:"error"`
}
