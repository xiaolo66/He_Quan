package okex

import (
	"He_Quan"
	"testing"
)

var rest = New(He_Quan.Options{AccessKey: "", SecretKey: "", PassPhrase: ""})
var orderID string

func TestOkexRest_FetchMarkets(t *testing.T) {
	if _, err := rest.FetchMarkets(); err != nil {
		t.Error(err)
	}
}

func TestOkexRest_FetchOrderBook(t *testing.T) {
	orderBook, err := rest.FetchOrderBook(symbol, 50)
	if err != nil {
		t.Error(err)
	}
	t.Log(orderBook)
}

func TestOkexRest_FetchTicker(t *testing.T) {
	ticker, err := rest.FetchTicker(symbol)
	if err != nil {
		t.Error(err)
	}
	t.Log(ticker)
}

func TestOkexRest_FetchAllTicker(t *testing.T) {
	tickers, err := rest.FetchAllTicker()
	if err != nil {
		t.Error(err)
	}
	t.Log(tickers)
}

func TestOkexRest_FetchTrade(t *testing.T) {
	trade, err := rest.FetchTrade(symbol)
	if err != nil {
		t.Error(err)
	}
	t.Log(trade)
}

func TestOkexRest_FetchKLine(t *testing.T) {
	klines, err := rest.FetchKLine(symbol, He_Quan.KLine1Minute)
	if err != nil {
		t.Error(err)
	}
	t.Log(klines)
}

func TestOkexRest_FetchBalance(t *testing.T) {
	balances, err := rest.FetchBalance()
	if err != nil {
		t.Error(err)
	}
	t.Log(balances)
}

func TestOkexRest_CreateOrder(t *testing.T) {
	order, err := rest.CreateOrder(symbol, 3000, 0.001, He_Quan.Buy, He_Quan.LIMIT, He_Quan.PostOnly, false)
	if err != nil {
		t.Error(err)
	}
	t.Log(order)
}

func TestOkexRest_FetchOrder(t *testing.T) {
	order, err := rest.FetchOrder(symbol, "6938997229316096")
	if err != nil {
		t.Error(err)
	}
	t.Log(order)
}

func TestOkexRest_FetchOpenOrders(t *testing.T) {
	orders, err := rest.FetchOpenOrders(symbol, 1, 10)
	if err != nil {
		t.Error(err)
	}
	t.Log(orders)
}

func TestOkexRest_CancelOrder(t *testing.T) {
	//order, err := rest.CreateOrder(symbol, 10000, 0.001, He_Quan.Buy, He_Quan.LIMIT, He_Quan.Normal, false)
	err := rest.CancelOrder(symbol, "6938997229316096")
	if err != nil {
		t.Error(err)
	}
}

