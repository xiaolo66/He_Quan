package binance
import (
	"github.com/xiaolo66/He_Quan"
	"fmt"
	"strconv"
	"strings"
	"time"
	."github.com/xiaolo66/He_Quan/utils"


)

// ResponseEvent 解析回包中的事件
type ResponseEvent struct {
	Event  string `json:"e"`
	Ignore int    `json:"E"` // unmarshal不区分大小写，会把E值解析到e上
}

type Filter struct {
	FilterType string `json:"filterType"`
	TickSize   string `json:"tickSize"`
	StepSize   string `json:"stepSize"`
}
type Market struct {
	Symbol             string   `json:"symbol"`
	Status             string   `json:"status"`
	BaseAsset          string   `json:"baseAsset"`
	BaseAssetPrecision int      `json:"baseAssetPrecision"`
	QuoteAsset         string   `json:"quoteAsset"`
	QuotePrecision     int      `json:"quotePrecision"`
	Filters            []Filter `json:"filters"`
	Contractype        string   `json:"contractType"`
}

type ExchangeInfo struct {
	Markets []Market `json:"symbols"`
}

// RawOrderBook
type RawOrderBook struct {
	Event         string        `json:"e"`                     // Event type
	EventTime     time.Duration `json:"E"`                     // Event time
	Symbol        string        `json:"s"`                     // Symbol ID
	FirstUpdateID int64         `json:"U"`                     // First update ID in event
	LastUpdateID  int64         `json:"u" rest:"lastUpdateId"` // Final update ID in event
	PreUpdateID   int64         `json:"pu"`
	Bids          He_Quan.RawDepth `json:"b" rest:"bids"` // Bids to be updated
	Asks          He_Quan.RawDepth `json:"a" rest:"asks"` // Asks to be updated
}

// OrderBook
type OrderBook struct {
	LastUpdateID int64 `json:"lastUpdateId"` // Last update ID
	He_Quan.OrderBook
}

func (o *OrderBook) update(bookData RawOrderBook) {
	o.Bids = o.Bids.Update(bookData.Bids, true)
	o.Asks = o.Asks.Update(bookData.Asks, false)
	o.LastUpdateID = bookData.LastUpdateID
}

// OrderBook of one symbol
type SymbolOrderBook map[string]*OrderBook

type Ticker struct {
	Timestamp   float64 `json:"E" rest:"openTime"`
	Symbol      string  `json:"s" rest:"symbol"`
	Close       string  `json:"c" rest:"lastPrice"`
	Open        string  `json:"o" rest:"openPrice"`
	High        string  `json:"h" rest:"highPrice"`
	Low         string  `json:"l" rest:"lowPrice"`
	Vol         string  `json:"v" rest:"volume"`
	BestBid     string  `json:"b" rest:"bidPrice"`
	BestBidSize string  `json:"B"`
	BestAsk     string  `json:"a" rest:"askPrice"`
	BestAskSize string  `json:"A"`
	EIgnore     string  `json:"e"`
	CIgnore     float64 `json:"C"`
	LIgnore     float64 `json:"L"`
	OIgnore     float64 `json:"O"`
}

func (t Ticker) parseTicker(symbol string) He_Quan.Ticker {
	return He_Quan.Ticker{
		Symbol:         symbol,
		Timestamp:      time.Duration(t.Timestamp),
		Open:           SafeParseFloat(t.Open),
		Last:           SafeParseFloat(t.Close),
		High:           SafeParseFloat(t.High),
		Low:            SafeParseFloat(t.Low),
		Vol:            SafeParseFloat(t.Vol),
		BestBuyPrice:   SafeParseFloat(t.BestBid),
		BestBuyAmount:  SafeParseFloat(t.BestBidSize),
		BestSellPrice:  SafeParseFloat(t.BestAsk),
		BestSellAmount: SafeParseFloat(t.BestAskSize),
	}
}

type Trade struct {
	Timestamp float64 `json:"T"`
	Symbol    string  `json:"s"`
	Price     string  `json:"p"`
	Size      string  `json:"q"`
	IsSell    bool    `json:"m"`
	MIgnore   bool    `json:"M"`
}

func (t Trade) parseTrade(symbol string) He_Quan.Trade {
	trade := He_Quan.Trade{
		Symbol:    symbol,
		Timestamp: time.Duration(t.Timestamp),
		Price:     SafeParseFloat(t.Price),
		Amount:    SafeParseFloat(t.Size),
		Side:      He_Quan.Buy,
	}
	if t.IsSell {
		trade.Side = He_Quan.Sell
	}
	return trade
}

type KLine struct {
	Symbol string `json:"s"`
	Line   struct {
		BTimestamp float64 `json:"t"`
		ETimestamp float64 `json:"T"`
		Open       string  `json:"o"`
		Close      string  `json:"c"`
		High       string  `json:"h"`
		Low        string  `json:"l"`
		Volume     string  `json:"v"`
		LIgnore    float64 `json:"L"`
	} `json:"k"`
}

func parseKLienType(t He_Quan.KLineType) string {
	kt := ""
	switch t {
	case He_Quan.KLine1Minute:
		kt = "1m"
	case He_Quan.KLine3Minute:
		kt = "3m"
	case He_Quan.KLine5Minute:
		kt = "5m"
	case He_Quan.KLine15Minute:
		kt = "15m"
	case He_Quan.KLine30Minute:
		kt = "30m"
	case He_Quan.KLine1Hour:
		kt = "1h"
	case He_Quan.KLine2Hour:
		kt = "2h"
	case He_Quan.KLine4Hour:
		kt = "4h"
	case He_Quan.KLine6Hour:
		kt = "6h"
	case He_Quan.KLine8Hour:
		kt = "8h"
	case He_Quan.KLine12Hour:
		kt = "12h"
	case He_Quan.KLine1Day:
		kt = "1d"
	case He_Quan.KLine3Day:
		kt = "3d"
	case He_Quan.KLine1Week:
		kt = "1w"
	case He_Quan.KLine1Month:
		kt = "1M"
	}
	return kt
}

type Order struct {
	Event           string        `json:"e" `                                                    //Event type
	EventTime       time.Duration `json:"E" `                                                    //Event time
	ClientID        string        `json:"c" fj:"c"  rest:"clientOrderId" future:"clientOrderId"` //Client order ID
	ID              int64         `json:"i" fj:"i"  rest:"orderId"       future:"orderId"`       //Order ID
	Type            string        `json:"o" fj:"o"  rest:""              future:"origType"`      //(LIMIT...)
	Amount          string        `json:"q" fj:"q"  rest:"origQty"       future:"origQty"`       //
	Price           string        `json:"p" fj:"p"  rest:"price"         future:"price"`         //
	AvePrice        string        `json:"-" fj:"ap"`
	Filled          string        `json:"z" fj:"z"  rest:"executedQty"   future:"executedQty"`    //filled amount
	Cost            string        `json:"Z"         rest:"cummulativeQuoteQty" future:"cumQuote"` //filled money
	Symbol          string        `json:"s" fj:"s"  rest:"symbol"`                                //Symbol
	Side            string        `json:"S" fj:"S"  rest:"side"          future:"side"`           //(BUY, SELL)
	CreateTime      time.Duration `json:"O"         rest:"time"          future:"time"`           //creation time
	TransactionTime time.Duration `json:"T" fj:"T"  rest:"updateTime"    future:"updateTime"`     //Transaction time
	Status          string        `json:"X" fj:"X"  rest:"status"        future:"status"`         //(NEW,CANCELED,TRADE,EXPIRED,REJECTED)
	Positionside    string        `         fj:"ps"                      future:"positionSide"`
	CIgnore         string        `json:"C" fj:"C"`
	XIgnore         string        `json:"x" fj:"x"`
	IIgnore         int           `json:"I" fj:"I"`
	PIgnore         string        `json:"P" fj:"P"`
	QIgnore         string        `json:"Q" fj:"Q"`
	TIgnore         int           `json:"t" fj:"t"`
}

func (o Order) parseOrder(symbol string) He_Quan.Order {
	order := He_Quan.Order{
		ID:              fmt.Sprintf("%v", o.ID),
		ClientID:        o.ClientID,
		Symbol:          symbol,
		Price:           o.Price,
		Amount:          o.Amount,
		Filled:          o.Filled,
		Cost:            o.Cost,
		Type:            "",
		OrderType:       0,
		CreateTime:      o.CreateTime,
		TransactionTime: o.TransactionTime,
	}
	switch o.Side {
	case "BUY":
		if o.Positionside == "LONG" {
			order.Side = He_Quan.OpenLong
		} else if o.Positionside == "SHORT" {
			order.Side = He_Quan.CloseShort
		} else {
			order.Side = He_Quan.Buy
		}
	case "SELL":
		if o.Positionside == "LONG" {
			order.Side = He_Quan.CloseLong
		} else if o.Positionside == "SHORT" {
			order.Side = He_Quan.OpenShort
		} else {
			order.Side = He_Quan.Sell
		}
	}
	switch o.Type {
	case "LIMIT":
		order.Type = He_Quan.LIMIT
	case "MARKET":
		order.Type = He_Quan.MARKET
	}
	switch o.Status {
	case "NEW":
		order.Status = He_Quan.Open
	case "CANCELED":
		order.Status = He_Quan.Canceled
		filled, err := strconv.ParseFloat(order.Filled, 64)
		if err == nil && filled > 0 {
			order.Status = He_Quan.Close
		}
	case "PARTIALLY_FILLED":
		order.Status = He_Quan.Partial
	case "FILLED":
		order.Status = He_Quan.Close
	default:
		order.Status = He_Quan.OrderStatusUnKnown
	}
	return order
}

type Balance struct {
	Currency  string `json:"a" rest:"asset" future:"asset"`
	Available string `json:"f" rest:"free" future:"availableBalance"`
	Frozen    string `json:"l" rest:"locked"`
	Total     string `future:"crossWalletBalance"`
}

func (b Balance) parseBalance() He_Quan.Balance {
	return He_Quan.Balance{
		Asset:     strings.ToUpper(b.Currency),
		Available: SafeParseFloat(b.Available),
		Frozen:    SafeParseFloat(b.Frozen),
	}
}

type Balances struct {
	Timestamp float64   `json:"u" rest:"updateTime"`
	Balances  []Balance `json:"B" rest:"balances"`
}

type FuturePosition struct {
	Symbol                 string `json:"symbol"`
	AvgPrice               string `json:"entryPrice"`             //开仓均价
	InitialMargin          string `json:"initialMargin"`          //当前所需起始保证金
	PositionInitialMargin  string `json:"positionInitialMargin"`  //持仓所需起始保证金
	OpenOrderInitialMargin string `json:"openOrderInitialMargin"` //当前挂单所需起始保证金
	Margin                 string `json:"maintMargin"`            //维持保证金
	Isolated               bool   `json:"isolated"`               //逐仓，全仓
	Amount                 string `json:"positionAmt"`            //仓位数量
	Side                   string `json:"positionSide"`           //开多，开空
	Leverage               string `json:"leverage"`
}

func (f *FuturePosition) ParserFuturePosition(coin, symbol string) (positions He_Quan.FuturePositons) {
	positions.Coin = coin
	positions.Symbol = symbol
	positions.AvgPrice = f.AvgPrice
	positions.Margin = f.Margin
	positions.Amount = SafeParseFloat(f.Amount)
	positions.Leverage, _ = strconv.Atoi(f.Leverage)
	if f.Isolated {
		positions.MarginMode = He_Quan.FixedMargin
	} else {
		positions.MarginMode = He_Quan.CrossedMargin

	}
	switch f.Side {
	case "LONG":
		positions.PositionType = He_Quan.PositionLong
	case "SHORT":
		positions.PositionType = He_Quan.PositionShort
	}
	return
}

//asset info
type FutureAssetInfo struct {
	AssetName        string `json:"asset"`
	MarginBalance    string `json:"marginBalance"`          // = Available + Freeze + AllUnrealizedPnl
	Available        string `json:"availableBalance"`       // available balance amount
	Freeze           string `json:"initialMargin"`          // = PositionMargin + OpenOrderMargin
	PositionMargin   string `json:"positionInitialMargin"`  // frozen by position margin
	OpenOrderMargin  string `json:"openOrderInitialMargin"` // frozen by open order margin
	AllUnrealizedPnl string `json:"unrealizedProfit"`       // unrealized profit
}

type FutureAccount struct {
	MarginBalance    string `json:"totalMarginBalance"`          // = Available + Freeze + AllUnrealizedPnl
	Available        string `json:"availableBalance"`            // available balance amount
	Freeze           string `json:"totalInitialMargin"`          // = PositionMargin + OpenOrderMargin
	PositionMargin   string `json:"totalPositionInitialMargin"`  // frozen by position margin
	OpenOrderMargin  string `json:"totalOpenOrderInitialMargin"` // frozen by open order margin
	AllUnrealizedPnl string `json:"totalUnrealizedProfit"`       // unrealized profit
}

//account info
type FutureAccountInfo struct {
	FutureAccount
	Assets    []FutureAssetInfo `json:"assets"`
	Positions []FuturePosition  `json:"positions"`
}

func (f FutureAccountInfo) parseAccountInfo() (accountInfo He_Quan.FutureAccountInfo) {
	accountInfo.Account.Available = SafeParseFloat(f.Available)
	accountInfo.Account.Total = SafeParseFloat(f.MarginBalance)
	accountInfo.Account.Freeze = SafeParseFloat(f.Freeze)
	accountInfo.Account.AllUnrealizedPnl = SafeParseFloat(f.AllUnrealizedPnl)
	accountInfo.Account.PositionMargin = SafeParseFloat(f.PositionMargin)
	accountInfo.Account.OpenOrderMargin = SafeParseFloat(f.OpenOrderMargin)
	accountInfo.Assets = make(map[string]He_Quan.FutureAsset)
	for _, ele := range f.Assets {
		asset := He_Quan.FutureAsset{
			AssetName:        strings.ToUpper(ele.AssetName),
			Available:        SafeParseFloat(ele.Available),
			Freeze:           SafeParseFloat(ele.Freeze),
			AllUnrealizedPnl: SafeParseFloat(ele.AllUnrealizedPnl),
			Total:            SafeParseFloat(ele.MarginBalance),
			PositionMargin:   SafeParseFloat(ele.PositionMargin),
			OpenOrderMargin:  SafeParseFloat(ele.OpenOrderMargin),
		}
		accountInfo.Assets[ele.AssetName] = asset
	}
	return
}

type WsBP struct {
	WsB []WsBalance  `json:"B"`
	WsP []WsPosition `json:"P"`
}
type WsBalance struct {
	Currency  string `json:"a"`
	Available string `json:"cw"`
	Frozen    string
	Total     string `json:"wb"`
}

func (w *WsBalance) parserWsBalance() He_Quan.Balance {
	return He_Quan.Balance{
		Asset:     w.Currency,
		Available: SafeParseFloat(w.Available),
		Frozen:    SafeParseFloat(w.Total) - SafeParseFloat(w.Available),
	}
}

type WsPosition struct {
	Symbol     string `json:"s"`
	Amount     string `json:"pa"` //仓位数量
	AvgPrice   string `json:"ep"` //开仓均价
	Margin     string `json:"iw"` //保证金
	MarginMode string `json:"mt"` //逐仓，全仓
	Side       string `json:"ps"` //开多，开空
}

func (w *WsPosition) parserWsPosition(symbol string) He_Quan.FuturePositons {
	future := He_Quan.FuturePositons{
		Symbol:   symbol,
		AvgPrice: w.AvgPrice,
		Margin:   w.Margin,
		Amount:   SafeParseFloat(w.Amount),
	}

	switch w.MarginMode {
	case "isolated":
		future.MarginMode = He_Quan.FixedMargin
	default:
		future.MarginMode = He_Quan.CrossedMargin
	}
	switch w.Side {
	case "LONG":
		future.PositionType = He_Quan.PositionLong
	case "SHORT":
		future.PositionType = He_Quan.PositionShort
	default:
		future.PositionType = He_Quan.PositionTypeUnKonwn
	}

	return future
}

type WsBalances struct {
	Timestamp float64 `json:"T"`
	Event     WsBP    `json:"a"`
}

type MarkFundingRate struct {
	Symbol          string `json:"symbol" ws:"s"`
	MarkPrice       string `json:"markPrice" ws:"p"`
	IndexPrice      string `json:"indexPrice" ws:"i"`
	LastFundingRate string `json:"lastFundingRate" ws:"r"`
	NextFundingTime int    `json:"nextFundingTime"`
	PIgnore         string `ws:"P"`
}

func (m *MarkFundingRate) parserMarkPrice(symbol string) He_Quan.MarkPrice {
	return He_Quan.MarkPrice{
		Symbol: symbol,
		Price:  m.MarkPrice,
	}
}

type DualSidePosition struct {
	DaulSide bool `json:"dualSidePosition"`
}
