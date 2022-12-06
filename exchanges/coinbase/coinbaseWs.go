package coinbase

import (
	"He_Quan"
	"He_Quan/exchanges"
	"He_Quan/exchanges/websocket"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Stream map[string]string

type SubTopic struct {
	Topic        string
	Symbol       string
	MessageType  He_Quan.MessageType
	LastUpdateID float64
}

type CoinBaseWs struct {
	exchanges.BaseExchange
	orderBooks   map[string]*CoinBaseOrderBook
	subTopicInfo map[string]SubTopic
	errors       map[int]He_Quan.ExError
	loginLock    sync.Mutex
	loginChan    chan struct{}
	isLogin      bool
}

func (e *CoinBaseWs) Init(option He_Quan.Options) {
	e.BaseExchange.Init()
	e.Option = option
	e.orderBooks = make(map[string]*CoinBaseOrderBook)
	e.subTopicInfo = make(map[string]SubTopic)
	e.errors = map[int]He_Quan.ExError{}
	e.loginLock = sync.Mutex{}
	e.loginChan = make(chan struct{})
	e.isLogin = false
	if e.Option.WsHost == "" {
		e.Option.WsHost = "wss://ws-feed.pro.coinbase.com"
	}
}

func (e *CoinBaseWs) SubscribeOrderBook(symbol string, level, speed int, isIncremental bool, sub He_Quan.MessageChan) (string, error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return "", err
	}
	return e.subscribe(e.Option.WsHost, market.SymbolID, symbol, He_Quan.MsgOrderBook, false, sub)
}

func (e *CoinBaseWs) SubscribeTicker(symbol string, sub He_Quan.MessageChan) (string, error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return "", err
	}
	return e.subscribe(e.Option.WsHost, market.SymbolID, symbol, He_Quan.MsgTicker, false, sub)
}

func (e *CoinBaseWs) SubscribeTrades(symbol string, sub He_Quan.MessageChan) (string, error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return "", err
	}
	return e.subscribe(e.Option.WsHost, market.SymbolID, symbol, He_Quan.MsgTrade, false, sub)
}

func (e *CoinBaseWs) SubscribeAllTicker(sub He_Quan.MessageChan) (string, error) {
	return "", He_Quan.ExError{Code: He_Quan.NotImplement}
}

func (e *CoinBaseWs) SubscribeKLine(symbol string, t He_Quan.KLineType, sub He_Quan.MessageChan) (string, error) {
	return "", He_Quan.ExError{Code: He_Quan.NotImplement}
}

func (e *CoinBaseWs) UnSubscribe(event string, sub He_Quan.MessageChan) error {
	conn, err := e.ConnectionMgr.GetConnection(e.Option.WsHost, nil)
	if err != nil {
		return err
	}
	tuple := strings.Split(event, "#")
	data := make(map[string]interface{}, 3)
	if len(tuple) == 2 {
		data = map[string]interface{}{
			"product_ids": []string{tuple[0]},
			"type":        "subscribe",
			"channels":    []string{tuple[1]},
		}
	}
	if err := conn.SendJsonMessage(data); err != nil {
		return err
	}
	conn.UnSubscribe(sub)
	return nil
}

func (e *CoinBaseWs) SubscribeBalance(symbol string, sub He_Quan.MessageChan) (string, error) {
	return "", He_Quan.ExError{Code: He_Quan.NotImplement}
}

func (e *CoinBaseWs) SubscribeOrder(symbol string, sub He_Quan.MessageChan) (string, error) {
	return "", He_Quan.ExError{Code: He_Quan.NotImplement}
}

func (e *CoinBaseWs) Connect(url string) (*exchanges.Connection, error) {
	conn := exchanges.NewConnection()
	err := conn.Connect(
		websocket.SetExchangeName("CoinBase"),
		websocket.SetWsUrl(url),
		websocket.SetIsAutoReconnect(e.Option.AutoReconnect),
		websocket.SetEnableCompression(false),
		websocket.SetReadDeadLineTime(time.Minute),
		websocket.SetMessageHandler(e.messageHandler),
		websocket.SetErrorHandler(e.errorHandler),
		websocket.SetCloseHandler(e.closeHandler),
		websocket.SetReConnectedHandler(e.reConnectedHandler),
		websocket.SetDisConnectedHandler(e.disConnectedHandler),
	)
	return conn, err
}

func (e *CoinBaseWs) subscribe(url, topic, symbol string, t He_Quan.MessageType, needLogin bool, sub He_Quan.MessageChan) (string, error) {
	conn, err := e.ConnectionMgr.GetConnection(url, e.Connect)
	if err != nil {
		return "", err
	}
	var data map[string]interface{}
	if needLogin {

	} else {
		data = map[string]interface{}{
			"product_ids": []string{topic},
			"type":        "subscribe",
		}
		switch t {
		case He_Quan.MsgTicker:
			data["channels"] = []string{"ticker"}
			topic += "#ticker"
		case He_Quan.MsgTrade:
			data["channels"] = []string{"matches"}
			topic += "#matches"
		case He_Quan.MsgOrderBook:
			data["channels"] = []string{"level2"}
			topic += "#level2"
		}
	}

	if err := conn.SendJsonMessage(data); err != nil {
		return "", err
	}
	conn.Subscribe(sub)

	return topic, nil
}

func (e *CoinBaseWs) messageHandler(url string, message []byte) {
	res := Response{}
	if err := json.Unmarshal(message, &res); err != nil {
		e.errorHandler(url, fmt.Errorf("[huobiWs] messageHandler unmarshal error:%v", err))
		return
	}
	switch res.Type {
	case "ticker":
		e.handleTicker(url, message)
	case "match":
		e.handleTrade(url, message)
	case "snapshot", "l2update":
		e.handleDepth(url, message, res.Type)
	}
}

func (e *CoinBaseWs) reConnectedHandler(url string) {
	e.BaseExchange.ReConnectedHandler(url, nil)
}

func (e *CoinBaseWs) disConnectedHandler(url string, err error) {
	// clear cache data, Prevent getting dirty data
	e.BaseExchange.DisConnectedHandler(url, err, func() {
		e.isLogin = false
		delete(e.orderBooks, url)
	})
}

func (e *CoinBaseWs) closeHandler(url string) {
	// clear cache data and the connection
	e.BaseExchange.CloseHandler(url, func() {
		delete(e.orderBooks, url)
	})
}

func (e *CoinBaseWs) errorHandler(url string, err error) {
	e.BaseExchange.ErrorHandler(url, err, nil)
}

func (e *CoinBaseWs) handleTicker(url string, message []byte) {
	var data WsTickerRes
	if err := json.Unmarshal(message, &data); err != nil {
		e.errorHandler(url, fmt.Errorf("[coinBaseWs] handleTicker - message Unmarshal to ticker error:%v", err))
		return
	}
	market, err := e.GetMarketByID(data.Symbol)
	if err != nil {
		e.errorHandler(url, fmt.Errorf("[coinBaseWs] handleTicker - find market by id error:%v", err))
		return
	}
	ticker := data.parseTicker(market)
	e.ConnectionMgr.Publish(url, He_Quan.Message{Type: He_Quan.MsgTicker, Data: ticker})
}

func (e *CoinBaseWs) handleDepth(url string, message []byte, depthType string) {
	if depthType == "snapshot" {
		var data OrderBookRes
		if err := json.Unmarshal(message, &data); err != nil {
			e.errorHandler(url, fmt.Errorf("[huobiWs] handleDepth - message Unmarshal to ticker error:%v", err))
			return
		}
		market, err := e.GetMarketByID(data.Symbol)
		if err != nil {
			e.errorHandler(url, fmt.Errorf("[coinBaseWs] handleDepth - find market by id error:%v", err))
			return
		}
		symbolOrderBook, ok := e.orderBooks[market.Symbol]
		if !ok {
			symbolOrderBook = &CoinBaseOrderBook{}
			e.orderBooks[market.Symbol] = symbolOrderBook
		}
		var asks He_Quan.Depth = He_Quan.Depth{}
		for _, ask := range data.Asks {
			depthItem, err := ask.ParseRawDepthItem()
			if err != nil {
				e.errorHandler(url, fmt.Errorf("[coinBaseWs] handleDepth - parse depth item error:%v", err))
			}
			asks = append(asks, depthItem)
		}
		var bids He_Quan.Depth = He_Quan.Depth{}
		for _, bid := range data.Bids {
			depthItem, err := bid.ParseRawDepthItem()
			if err != nil {
				e.errorHandler(url, fmt.Errorf("[coinBaseWs] handleDepth - parse depth item error:%v", err))
			}
			bids = append(bids, depthItem)
		}
		sort.Sort(asks)
		sort.Sort(sort.Reverse(bids))
		symbolOrderBook.Asks = asks
		symbolOrderBook.Bids = bids
		symbolOrderBook.Symbol = market.Symbol
		e.ConnectionMgr.Publish(url, He_Quan.Message{Type: He_Quan.MsgOrderBook, Data: symbolOrderBook.OrderBook})
	} else {
		var data WsOrderBookUpdateRes
		if err := json.Unmarshal(message, &data); err != nil {
			e.errorHandler(url, fmt.Errorf("[huobiWs] handleDepth - message Unmarshal to ticker error:%v", err))
			return
		}
		market, err := e.GetMarketByID(data.Symbol)
		if err != nil {
			e.errorHandler(url, fmt.Errorf("[coinBaseWs] handleDepth - find market by id error:%v", err))
			return
		}
		symbolOrderBook, ok := e.orderBooks[market.Symbol]
		if !ok {
			e.errorHandler(url, fmt.Errorf("[coinBaseWs] handleDepth - find cache orderbook err:%v", err))
			return
		}
		var changeDepth OrderBookRes = OrderBookRes{
			Asks: He_Quan.RawDepth{},
			Bids: He_Quan.RawDepth{},
		}
		for _, change := range data.Changes {
			if change[0] == "sell" {
				changeDepth.Asks = append(changeDepth.Asks, He_Quan.RawDepthItem{
					change[1], change[2],
				})
			} else {
				changeDepth.Bids = append(changeDepth.Bids, He_Quan.RawDepthItem{
					change[1], change[2],
				})
			}
		}
		symbolOrderBook.update(changeDepth)

		e.ConnectionMgr.Publish(url, He_Quan.Message{Type: He_Quan.MsgOrderBook, Data: symbolOrderBook.OrderBook})
	}
}

func (e *CoinBaseWs) handleTrade(url string, message []byte) {
	var data WsTradeRes
	if err := json.Unmarshal(message, &data); err != nil {
		e.errorHandler(url, fmt.Errorf("[huobiWs] handleTicker - message Unmarshal to ticker error:%v", err))
		return
	}
	market, err := e.GetMarketByID(data.Symbol)
	if err != nil {
		e.errorHandler(url, fmt.Errorf("[coinBaseWs] handleTicker - find market by id error:%v", err))
		return
	}
	trade := data.parseTrade(market)
	e.ConnectionMgr.Publish(url, He_Quan.Message{Type: He_Quan.MsgTrade, Data: trade})
}
