package coinbase

import (
	"github.com/xiaolo66/He_Quan"
	"github.com/xiaolo66/He_Quan/exchanges"
	. "github.com/xiaolo66/He_Quan/utils"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type CoinBaseRest struct {
	exchanges.BaseExchange
}

var AccountId int = 0

func (e *CoinBaseRest) Init(option He_Quan.Options) {
	e.Option = option

	if e.Option.RestHost == "" {
		e.Option.RestHost = "https://api.pro.coinbase.com"
	}
	if e.Option.RestPrivateHost == "" {
		e.Option.RestPrivateHost = "https://api.pro.coinbase.com"
	}
}

func (e *CoinBaseRest) FetchOrderBook(symbol string, size int) (orderBook He_Quan.OrderBook, err error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return
	}
	params := url.Values{}
	params.Set("level", strconv.Itoa(size))
	res, err := e.Fetch(e, exchanges.Public, exchanges.GET, "/products/"+market.SymbolID+"/book", params, http.Header{})
	if err != nil {
		return
	}
	var orderBookRes OrderBookRes
	if err = json.Unmarshal(res, &orderBookRes); err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}
	orderBook = orderBookRes.parseOrderBook(symbol)
	return
}

func (e *CoinBaseRest) FetchTicker(symbol string) (ticker He_Quan.Ticker, err error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return
	}
	params := url.Values{}
	res, err := e.Fetch(e, exchanges.Public, exchanges.GET, "/products/"+market.SymbolID+"/stats", params, http.Header{})
	if err != nil {
		return
	}

	var data Stats24Hr
	if err = json.Unmarshal(res, &data); err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}
	params.Set("level", "1")
	res, err = e.Fetch(e, exchanges.Public, exchanges.GET, "/products/"+market.SymbolID+"/book", params, http.Header{})
	if err != nil {
		return
	}
	var orderBookRes OrderBookRes
	if err = json.Unmarshal(res, &orderBookRes); err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}
	bestBidsItem, err := orderBookRes.Bids[0].ParseRawDepthItem()
	if err != nil {
		return
	}
	bestAsksItem, err := orderBookRes.Asks[0].ParseRawDepthItem()
	if err != nil {
		return
	}
	ticker = He_Quan.Ticker{
		Symbol:         symbol,
		Timestamp:      time.Duration(time.Now().Unix()),
		BestBuyPrice:   SafeParseFloat(bestBidsItem.Price),
		BestSellPrice:  SafeParseFloat(bestAsksItem.Price),
		BestBuyAmount:  SafeParseFloat(bestBidsItem.Amount),
		BestSellAmount: SafeParseFloat(bestAsksItem.Amount),
		Open:           SafeParseFloat(data.Open),
		Last:           SafeParseFloat(data.Last),
		High:           SafeParseFloat(data.High),
		Low:            SafeParseFloat(data.Low),
		Vol:            SafeParseFloat(data.Volume),
	}
	return
}

func (e *CoinBaseRest) FetchAllTicker() (tickers map[string]He_Quan.Ticker, err error) {
	params := url.Values{}
	res, err := e.Fetch(e, exchanges.Public, exchanges.GET, "/products/stats", params, http.Header{})
	if err != nil {
		return
	}
	var data map[string]AllTickerItem
	if err = json.Unmarshal(res, &data); err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}
	tickers = make(map[string]He_Quan.Ticker)
	for symbol, t := range data {
		market, err := e.GetMarketByID(symbol)
		if err != nil {
			continue
		}
		tickers[market.Symbol] = He_Quan.Ticker{
			Open:   SafeParseFloat(t.Ticker.Open),
			Last:   SafeParseFloat(t.Ticker.Last),
			High:   SafeParseFloat(t.Ticker.High),
			Low:    SafeParseFloat(t.Ticker.Low),
			Vol:    SafeParseFloat(t.Ticker.Volume),
			Symbol: market.Symbol,
		}
	}
	return
}

func (e *CoinBaseRest) FetchTrade(symbol string) (trades []He_Quan.Trade, err error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return
	}
	params := url.Values{}
	res, err := e.Fetch(e, exchanges.Public, exchanges.GET, "/products/"+market.SymbolID+"/trades", params, http.Header{})
	if err != nil {
		return
	}
	//println(string(res))
	var data []Trade
	if err = json.Unmarshal(res, &data); err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}
	for _, t := range data {
		trade, err := t.parseTrade(symbol)
		if err == nil {
			trades = append(trades, trade)
		}
	}
	return
}

func (e *CoinBaseRest) FetchKLine(symbol string, t He_Quan.KLineType) (klines []He_Quan.KLine, err error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return
	}
	params := url.Values{}
	switch t {
	case He_Quan.KLine15Minute:
		params.Set("granularity", "900")
	case He_Quan.KLine1Day:
		params.Set("granularity", "86400")
	case He_Quan.KLine1Minute:
		params.Set("granularity", "60")
	case He_Quan.KLine1Hour:
		params.Set("granularity", "3600")
	case He_Quan.KLine6Hour:
		params.Set("granularity", "21600")
	case He_Quan.KLine5Minute:
		params.Set("granularity", "300")
	default:
		return nil, errors.New("coinbase can not support kline interval")
	}
	res, err := e.Fetch(e, exchanges.Public, exchanges.GET, "/products/"+market.SymbolID+"/candles", params, http.Header{})
	if err != nil {
		return
	}
	var data KlineRes
	if err = json.Unmarshal(res, &data); err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}
	klines = data.parseKLine(market, t)
	return
}

func (e *CoinBaseRest) FetchMarkets() (map[string]He_Quan.Market, error) {
	if len(e.Option.Markets) > 0 {
		return e.Option.Markets, nil
	}
	res, err := e.Fetch(e, exchanges.Public, exchanges.GET, "/products", url.Values{}, http.Header{})
	if err != nil {
		return e.Option.Markets, err
	}

	var markets SymbolListRes
	if err = json.Unmarshal(res, &markets); err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return e.Option.Markets, err
	}

	e.Option.Markets = make(map[string]He_Quan.Market)
	for _, value := range markets {
		market := He_Quan.Market{
			SymbolID: value.Symbol,
			Symbol:   strings.ToUpper(fmt.Sprintf("%v/%v", value.Base, value.Quote)),
			BaseID:   strings.ToUpper(value.Base),
			QuoteID:  strings.ToUpper(value.Quote),
			Lot:      SafeParseFloat(value.MinAmount),
		}
		pres := strings.Split(value.PricePrecision, ".")
		if len(pres) == 1 {
			market.PricePrecision = 0
		} else {
			market.PricePrecision = len(pres[1])
		}

		pres = strings.Split(value.AmountPrecision, ".")
		if len(pres) == 1 {
			market.AmountPrecision = 0
		} else {
			market.AmountPrecision = len(pres[1])
		}
		e.Option.Markets[market.Symbol] = market
	}
	return e.Option.Markets, nil
}
func (e *CoinBaseRest) FetchBalance() (balances map[string]He_Quan.Balance, err error) {
	err = He_Quan.ExError{Code: He_Quan.NotImplement}
	return
}

func (e *CoinBaseRest) CreateOrder(symbol string, price, amount float64, side He_Quan.Side, tradeType He_Quan.TradeType, orderType He_Quan.OrderType, useClientID bool) (order He_Quan.Order, err error) {
	err = He_Quan.ExError{Code: He_Quan.NotImplement}
	return
}

func (e *CoinBaseRest) CancelOrder(symbol, orderID string) (err error) {
	err = He_Quan.ExError{Code: He_Quan.NotImplement}
	return
}

func (e *CoinBaseRest) CancelAllOrders(symbol string) (err error) {
	err = He_Quan.ExError{Code: He_Quan.NotImplement}
	return
}

func (e *CoinBaseRest) FetchOrder(symbol, orderID string) (order He_Quan.Order, err error) {
	err = He_Quan.ExError{Code: He_Quan.NotImplement}
	return
}

func (e *CoinBaseRest) FetchOpenOrders(symbol string, pageIndex, pageSize int) (orders []He_Quan.Order, err error) {
	err = He_Quan.ExError{Code: He_Quan.NotImplement}
	return
}

func (e *CoinBaseRest) Sign(access, method, function string, param url.Values, header http.Header) (request exchanges.Request) {
	request.Headers = header
	request.Method = method
	if access == exchanges.Public {
		request.Url = fmt.Sprintf("%s%s", e.Option.RestHost, function)
		if len(param) > 0 {
			request.Url = request.Url + "?" + param.Encode()
		}
	} else {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		key, err := base64.StdEncoding.DecodeString(e.Option.SecretKey)
		if err != nil {
			return
		}
		signature := hmac.New(sha256.New, key)
		_, err = signature.Write([]byte(fmt.Sprintf(
			"%s%s%s%s",
			timestamp,
			method,
			function,
			UrlValuesToJson(param),
		)))
		if err != nil {
			return
		}
		request.Headers.Set("Content-Type", "application/json")
		request.Headers.Set("CB-ACCESS-KEY", e.Option.AccessKey)
		request.Headers.Set("CB-ACCESS-PASSPHRASE", e.Option.PassPhrase)
		request.Headers.Set("CB-ACCESS-TIMESTAMP", timestamp)
		request.Headers.Set("CB-ACCESS-SIGN", base64.StdEncoding.EncodeToString(signature.Sum(nil)))
		request.Url = fmt.Sprintf("%s%s", e.Option.RestPrivateHost, function)
		if len(param) > 0 {
			request.Url = request.Url + "?" + param.Encode()
		}
	}
	return request
}

func (e *CoinBaseRest) HandleError(request exchanges.Request, response []byte) error {
	type Result struct {
		Message string
	}
	var result Result
	if err := json.Unmarshal(response, &result); err != nil {
		return nil
	}
	if result.Message == "" {
		return nil
	}
	return He_Quan.ExError{Message: result.Message}
}
