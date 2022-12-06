package binance


import (
	"github.com/xiaolo66/He_Quan"
	"fmt"
	"testing"


)

var BaFuture = NewFuture(He_Quan.Options{
	PassPhrase: "",
	ProxyUrl:   "http://127.0.0.1:4780"}, He_Quan.FutureOptions{
	FutureAccountType: He_Quan.UsdtMargin,
	ContractType:      He_Quan.Swap,
})
var fmsgChan = make(He_Quan.MessageChan)

func handleFutureMsg(msgChan <-chan He_Quan.Message) {
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
				ticker, ok := msg.Data.(He_Quan.Ticker)
				if !ok {
					fmt.Printf("ticker data error %v", msg)
				}
				fmt.Printf("ticker:%+v\n", ticker)
			case He_Quan.MsgTrade:
				trades, ok := msg.Data.(He_Quan.Trade)
				if !ok {
					fmt.Printf("trade data error %v", msg)
				}
				fmt.Printf("trade:%+v\n", trades)
			case He_Quan.MsgMarkPrice:
				markPrice, ok := msg.Data.(He_Quan.MarkPrice)
				if !ok {
					fmt.Printf("markprice data error %v", msg)
				}
				fmt.Printf("markPrice:%+v\n", markPrice)
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
			case He_Quan.MsgPositions:
				position, ok := msg.Data.(He_Quan.FuturePositonsUpdate)
				if !ok {
					fmt.Printf("order data error %v", msg)
				}
				fmt.Printf("position:%+v\n", position)
			}
		}
	}
}

func TestBinanceFutureWs_SubscribeOrderBook(t *testing.T) {

	if _, err := BaFuture.SubscribeOrderBook(symbol, 5, 0, true, fmsgChan); err == nil {
		handleFutureMsg(fmsgChan)
	}
}

func TestBinanceFutureWs_SubscribeTicker(t *testing.T) {
	if _, err := BaFuture.SubscribeTicker(symbol, fmsgChan); err == nil {
		handleFutureMsg(fmsgChan)
	}
}

func TestBinanceFutureWs_SubscribeTrades(t *testing.T) {
	if _, err := BaFuture.SubscribeTrades(symbol, fmsgChan); err == nil {
		handleFutureMsg(fmsgChan)
	}
}

func TestBinanceFutureWs_SubscribeKLine(t *testing.T) {
	if _, err := BaFuture.SubscribeKLine(symbol, He_Quan.KLine3Minute, fmsgChan); err == nil {
		handleFutureMsg(fmsgChan)
	}
}

func TestBinanceFutureWs_SubscribeBalance(t *testing.T) {
	if _, err := BaFuture.SubscribeBalance(symbol, fmsgChan); err == nil {
		handleFutureMsg(fmsgChan)
	}
}
func TestZbFutureWs_SubscribeOrder(t *testing.T) {
	if _, err := BaFuture.SubscribeOrder(symbol, fmsgChan); err == nil {
		handleFutureMsg(fmsgChan)
	}
}

func TestBinanceFutureWs_SubscribePositions(t *testing.T) {
	if _, err := BaFuture.SubscribePositions(symbol, fmsgChan); err == nil {
		handleFutureMsg(fmsgChan)
	}
}

func TestBinanceFutureWs_SubscribeMarkPrice(t *testing.T) {
	if _, err := BaFuture.SubscribeMarkPrice("EOS/USDT", fmsgChan); err == nil {
		handleFutureMsg(fmsgChan)
	}
}
