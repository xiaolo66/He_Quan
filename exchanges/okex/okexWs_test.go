package okex

import (
	"github.com/xiaolo66/He_Quan"
	"fmt"
	"testing"
)

var (
	symbol  = "BTC/USDT"
	e       = New(He_Quan.Options{AccessKey: "", SecretKey: "", PassPhrase: "", AutoReconnect: true})
	msgChan = make(He_Quan.MessageChan)
)

func handleMsg(msgChan <-chan He_Quan.Message) {
	for {
		select {
		case msg := <-msgChan:
			switch msg.Type {
			case He_Quan.MsgReConnected:
				fmt.Println("reconnected..., restart subscribe")
			case He_Quan.MsgDisConnected:
				fmt.Println("disconnected, stop use old data, waiting reconnect....")
			case He_Quan.MsgClosed:
				fmt.Println("websocket closed, stop all")
				break
			case He_Quan.MsgError:
				fmt.Printf("error happend: %v\n", msg.Data)
				if err, ok := msg.Data.(He_Quan.ExError); ok {
					if err.Code == He_Quan.ErrInvalidDepth {
						//depth data invalid, Do some cleanup work, wait for the latest data or resubscribe
					}
				}
			case He_Quan.MsgOrderBook:
				orderbook, ok := msg.Data.(He_Quan.OrderBook)
				if !ok {
					fmt.Printf("order book data error %v", msg)
				}
				fmt.Printf("order book:%+v\n", orderbook)
			case He_Quan.MsgTicker:
				ticker, ok := msg.Data.([]He_Quan.Ticker)
				if !ok {
					fmt.Printf("ticker data error %v", msg)
				}
				fmt.Printf("ticker:%+v\n", ticker)
			case He_Quan.MsgTrade:
				trades, ok := msg.Data.([]He_Quan.Trade)
				if !ok {
					fmt.Printf("trade data error %v", msg)
				}
				fmt.Printf("trade:%+v\n", trades)
			case He_Quan.MsgKLine:
				klines, ok := msg.Data.([]He_Quan.KLine)
				if !ok {
					fmt.Printf("kline data error %v", msg)
				}
				fmt.Printf("kline:%+v\n", klines)
			case He_Quan.MsgBalance:
				balances, ok := msg.Data.(He_Quan.BalanceUpdate)
				if !ok {
					fmt.Printf("balance data error %v", msg)
				}
				fmt.Printf("balance:%+v\n", balances)
			case He_Quan.MsgOrder:
				order, ok := msg.Data.(He_Quan.Order)
				if !ok {
					fmt.Printf("order data error %v", msg)
				}
				fmt.Printf("order:%+v\n", order)
			}
		}
	}
}

func TestOkexWs_SubscribeOrderBook(t *testing.T) {
	if _, err := e.SubscribeOrderBook(symbol, 0, 0, true, msgChan); err == nil {
		handleMsg(msgChan)
	}
}

func TestOkexWs_SubscribeTicker(t *testing.T) {
	if _, err := e.SubscribeTicker(symbol, msgChan); err == nil {
		handleMsg(msgChan)
	}
}

func TestOkexWs_SubscribeTrades(t *testing.T) {
	if _, err := e.SubscribeTrades(symbol, msgChan); err == nil {
		handleMsg(msgChan)
	}
}

func TestOkexWs_SubscribeKLine(t *testing.T) {
	if _, err := e.SubscribeKLine(symbol, He_Quan.KLine1Minute, msgChan); err == nil {
		handleMsg(msgChan)
	}
}

func TestOkexWs_SubscribeBalance(t *testing.T) {
	if _, err := e.SubscribeBalance(symbol, msgChan); err == nil {
		handleMsg(msgChan)
	}
}
func TestOkexWs_SubscribeOrder(t *testing.T) {
	if _, err := e.SubscribeOrder(symbol, msgChan); err == nil {
		handleMsg(msgChan)
	}
}
