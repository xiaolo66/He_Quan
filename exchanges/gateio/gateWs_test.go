package gateio

import (
	"He_Quan"
	"fmt"
	"testing"
)

var symbols = "EOS/USDT"

var e = New(He_Quan.Options{
	ExchangeName: "gate",
	SecretKey:    "",
	AccessKey:    "",
	ProxyUrl:     "http://127.0.0.1:4780",
})

var msgChan = make(He_Quan.MessageChan)

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
						// symbol := err.Data["symbol"]
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
				ticker, ok := msg.Data.(He_Quan.Ticker)
				if !ok {
					fmt.Printf("ticker data error %v", msg)
				}
				fmt.Printf("ticker:%+v\n", ticker)
			case He_Quan.MsgTrade:
				trade, ok := msg.Data.(He_Quan.Trade)
				if !ok {
					fmt.Printf("trade data error %v", msg)
				}
				fmt.Printf("trade:%+v\n", trade)
			case He_Quan.MsgKLine:
				klines, ok := msg.Data.(He_Quan.KLine)
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

func TestGateWs_SubscribeOrderBook(t *testing.T) {
	_, err := e.SubscribeOrderBook(symbols, 200, 1000, false, msgChan)
	if err == nil {
		handleMsg(msgChan)
	}
}

func TestGateWs_SubscribeTicker(t *testing.T) {
	_, err := e.SubscribeTicker(symbols, msgChan)
	if err == nil {
		handleMsg(msgChan)
	}
}

func TestGateWs_SubscribeTrades(t *testing.T) {
	_, err := e.SubscribeTrades(symbols, msgChan)
	if err == nil {
		handleMsg(msgChan)
	}
}



func TestGateWs_SubscribeKLine(t *testing.T) {
	_, err := e.SubscribeKLine(symbols,He_Quan.KLine1Minute ,msgChan)
	if err == nil {
		handleMsg(msgChan)
	}
}

func TestGateWs_SubscribeBalance(t *testing.T) {
	_, err := e.SubscribeBalance(symbol, msgChan)
	if err == nil {
		handleMsg(msgChan)
	}
}

func TestGateWs_SubscribeOrder(t *testing.T) {
	_, err := e.SubscribeOrder(symbol, msgChan)
	if err == nil {
		handleMsg(msgChan)
	}
}

func TestGateWs_getSnapshotOrderBook(t *testing.T){
	var sym =make(SymbolOrderBook)
	err:=e.getSnapshotOrderBook("2",He_Quan.Market{
		SymbolID: "EOS_USDT",
		Symbol: symbol,
	},&sym)
	if err!=nil{
		fmt.Println(err)
	}
	fmt.Println((*e.orderBooks["2"])[symbol].OrderBook)

}

