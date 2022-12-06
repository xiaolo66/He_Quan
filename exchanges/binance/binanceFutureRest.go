package binance

import (
	"He_Quan"
	"He_Quan/exchanges"
	"He_Quan/utils"
	"encoding/json"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type BinanceFutureRest struct {
	accountType  He_Quan.FutureAccountType
	contractType He_Quan.ContractType
	futuresKind  He_Quan.FuturesKind

	exchanges.BaseExchange
	errors map[int]RawError
}

func (e *BinanceFutureRest) Init(option He_Quan.Options) {
	e.Option = option
	e.errors = make(map[int]RawError)

	if e.Option.RestHost == "" {
		e.Option.RestHost = "https://fapi.binance.com"
	}
	if e.Option.RestPrivateHost == "" {
		e.Option.RestPrivateHost = "https://fapi.binance.com"
	}
}

func (e *BinanceFutureRest) FetchOrderBook(symbol string, size int) (orderBook He_Quan.OrderBook, err error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return
	}
	params := url.Values{}
	params.Set("symbol", market.SymbolID)
	params.Set("limit", strconv.Itoa(size))
	res, err := e.Fetch(e, exchanges.Public, exchanges.GET, "/fapi/v1/depth", params, http.Header{})
	if err != nil {
		return
	}
	var data RawOrderBook
	restJson := jsoniter.Config{TagKey: "rest"}.Froze()
	err = restJson.Unmarshal(res, &data)
	if err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}
	ob := OrderBook{}
	ob.update(data)
	return ob.OrderBook, nil
}

func (e *BinanceFutureRest) FetchTicker(symbol string) (ticker He_Quan.Ticker, err error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return
	}
	params := url.Values{}
	params.Set("symbol", market.SymbolID)
	res, err := e.Fetch(e, exchanges.Public, exchanges.GET, "/fapi/v1/ticker/24hr", params, http.Header{})
	if err != nil {
		return
	}

	var data Ticker
	restJson := jsoniter.Config{TagKey: "rest"}.Froze()
	if err = restJson.Unmarshal(res, &data); err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}
	ticker = data.parseTicker(market.Symbol)
	return
}
func (e *BinanceFutureRest) FetchAllTicker() (tickers map[string]He_Quan.Ticker, err error) {
	params := url.Values{}
	res, err := e.Fetch(e, exchanges.Public, exchanges.GET, "/fapi/v1/ticker/price", params, http.Header{})
	if err != nil {
		return
	}

	var data = make([]struct {
		Symbol string  `json:"symbol"`
		Price  string  `json:"price"`
		Time   float64 `json:"time"`
	}, 0)
	if err = json.Unmarshal(res, &data); err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}

	tickers = make(map[string]He_Quan.Ticker, 0)
	for _, t := range data {
		market, err := e.GetMarketByID(t.Symbol)
		if err != nil {
			continue
		}
		tickers[market.Symbol] = He_Quan.Ticker{Symbol: market.Symbol, Last: utils.SafeParseFloat(t.Price), Timestamp: time.Duration(t.Time)}
	}

	return
}

func (e *BinanceFutureRest) FetchTrade(symbol string) (trades []He_Quan.Trade, err error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return
	}
	params := url.Values{}
	params.Set("symbol", market.SymbolID)
	res, err := e.Fetch(e, exchanges.Public, exchanges.GET, "/fapi/v1/aggTrades", params, http.Header{})
	if err != nil {
		return
	}

	var data []Trade
	if err = json.Unmarshal(res, &data); err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}

	for _, t := range data {
		trades = append(trades, t.parseTrade(market.Symbol))
	}
	return
}

func (e *BinanceFutureRest) FetchKLine(symbol string, t He_Quan.KLineType) (klines []He_Quan.KLine, err error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return
	}
	params := url.Values{}
	params.Set("symbol", market.SymbolID)
	params.Set("interval", parseKLienType(t))
	res, err := e.Fetch(e, exchanges.Public, exchanges.GET, "/fapi/v1/klines", params, http.Header{})
	if err != nil {
		return
	}

	var data [][]interface{}
	if err = json.Unmarshal(res, &data); err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}

	for _, ele := range data {
		if len(ele) < 6 {
			continue
		}
		var (
			timestamp                  float64
			open, last, high, low, vol string
		)
		utils.SafeAssign(ele[0], &timestamp)
		utils.SafeAssign(ele[1], &open)
		utils.SafeAssign(ele[2], &high)
		utils.SafeAssign(ele[3], &low)
		utils.SafeAssign(ele[4], &last)
		utils.SafeAssign(ele[5], &vol)
		kline := He_Quan.KLine{
			Symbol:    market.Symbol,
			Type:      t,
			Timestamp: time.Duration(timestamp),
			Open:      utils.SafeParseFloat(open),
			High:      utils.SafeParseFloat(high),
			Low:       utils.SafeParseFloat(low),
			Close:     utils.SafeParseFloat(last),
			Volume:    utils.SafeParseFloat(vol),
		}
		klines = append([]He_Quan.KLine{kline}, klines...)
	}
	return
}

func (e *BinanceFutureRest) FetchMarkets() (map[string]He_Quan.Market, error) {
	if len(e.Option.Markets) > 0 {
		return e.Option.Markets, nil
	}
	res, err := e.Fetch(e, exchanges.Public, exchanges.GET, "/fapi/v1/exchangeInfo", url.Values{}, http.Header{})
	if err != nil {
		return e.Option.Markets, err
	}
	e.Option.Markets = make(map[string]He_Quan.Market)
	var info ExchangeInfo
	if err := json.Unmarshal(res, &info); err != nil {
		return e.Option.Markets, err
	}
	for _, m := range info.Markets {
		if m.Status != "TRADING" {
			continue
		}
		pricePrecision := m.QuotePrecision
		amountPrecision := m.BaseAssetPrecision
		for _, filter := range m.Filters {
			if filter.FilterType == "PRICE_FILTER" {
				re, _ := regexp.Compile(`0+$`)
				s := re.ReplaceAllString(filter.TickSize, "")
				pres := strings.Split(s, ".")
				if len(pres) == 2 {
					pricePrecision = len(pres[1])
				}
			}
			if filter.FilterType == "LOT_SIZE" {
				re, _ := regexp.Compile(`0+$`)
				s := re.ReplaceAllString(filter.StepSize, "")
				pres := strings.Split(s, ".")
				if len(pres) == 2 {
					amountPrecision = len(pres[1])
				}
			}
		}
		if e.contractType == He_Quan.Swap {
			if m.Contractype == "CURRENT_QUARTER" {
				continue
			}
			market := He_Quan.Market{
				SymbolID:        strings.ToUpper(m.Symbol),
				Symbol:          strings.ToUpper(fmt.Sprintf("%s/%s", m.BaseAsset, m.QuoteAsset)),
				BaseID:          strings.ToUpper(m.BaseAsset),
				QuoteID:         strings.ToUpper(m.QuoteAsset),
				PricePrecision:  pricePrecision,
				AmountPrecision: amountPrecision,
			}
			e.Option.Markets[market.Symbol] = market
		}
		if e.contractType == He_Quan.Futures {
			if m.Contractype != "CURRENT_QUARTER" {
				continue
			}
			market := He_Quan.Market{
				SymbolID:        strings.ToUpper(m.Symbol),
				Symbol:          strings.ToUpper(fmt.Sprintf("%s/%s", m.BaseAsset, m.QuoteAsset)),
				BaseID:          strings.ToUpper(m.BaseAsset),
				QuoteID:         strings.ToUpper(m.QuoteAsset),
				PricePrecision:  pricePrecision,
				AmountPrecision: amountPrecision,
			}
			e.Option.Markets[market.Symbol] = market
		}
	}
	return e.Option.Markets, nil
}

func (e *BinanceFutureRest) CreateOrder(symbol string, price, amount float64, side He_Quan.Side, tradeType He_Quan.TradeType, orderType He_Quan.OrderType, useClientID bool) (order He_Quan.Order, err error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return
	}
	params := url.Values{}
	params.Set("symbol", market.SymbolID)
	params.Set("quantity", utils.Round(amount, market.AmountPrecision, false))
	switch side {
	case He_Quan.OpenLong:
		params.Set("side", "BUY")
		params.Set("positionSide", "LONG")
	case He_Quan.OpenShort:
		params.Set("side", "SELL")
		params.Set("positionSide", "SHORT")
	case He_Quan.CloseLong:
		params.Set("side", "SELL")
		params.Set("positionSide", "LONG")
	case He_Quan.CloseShort:
		params.Set("side", "BUY")
		params.Set("positionSide", "SHORT")
	}
	switch tradeType {
	case He_Quan.LIMIT:
		params.Set("type", "LIMIT")
		params.Set("price", utils.Round(price, market.PricePrecision, false))
		params.Set("timeInForce", "GTC")
	case He_Quan.MARKET:
		params.Set("type", "MARKET")
	}
	if useClientID {
		params.Set("newClientOrderId", utils.GenerateOrderClientId(e.Option.ClientOrderIDPrefix, 32))
	}
	params.Set("newOrderRespType", "ACK")
	res, err := e.Fetch(e, exchanges.Private, exchanges.POST, "/fapi/v1/order", params, http.Header{})
	if err != nil {
		return
	}
	fmt.Println(string(res))
	type response struct {
		ID  int64  `json:"orderId"`
		CID string `json:"clientOrderId"`
	}
	var data response
	err = json.Unmarshal(res, &data)
	if err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}
	order.ID = strconv.FormatInt(data.ID, 10)
	order.ClientID = data.CID
	return
}

func (e *BinanceFutureRest) CancelOrder(symbol, orderID string) (err error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return
	}
	params := url.Values{}
	if utils.IsClientOrderID(orderID, e.Option.ClientOrderIDPrefix) {
		params.Set("clientOrderId", orderID)
	} else {
		params.Set("orderId", orderID)
	}
	params.Set("symbol", market.SymbolID)
	_, err = e.Fetch(e, exchanges.Private, exchanges.DELETE, "/fapi/v1/order", params, http.Header{})
	return err
}

func (e *BinanceFutureRest) CancelAllOrders(symbol string) (err error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return
	}
	params := url.Values{}
	params.Set("symbol", market.SymbolID)
	_, err = e.Fetch(e, exchanges.Private, exchanges.DELETE, "/fapi/v1/allOpenOrders", params, http.Header{})

	return err
}

func (e *BinanceFutureRest) FetchOrder(symbol, orderID string) (order He_Quan.Order, err error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return
	}
	params := url.Values{}
	params.Set("symbol", market.SymbolID)
	if utils.IsClientOrderID(orderID, e.Option.ClientOrderIDPrefix) {
		params.Set("origClientOrderId", orderID)
	} else {
		params.Set("orderId", orderID)
	}
	res, err := e.Fetch(e, exchanges.Private, exchanges.GET, "/fapi/v1/order", params, http.Header{})
	if err != nil {
		return
	}
	var data Order
	restJson := jsoniter.Config{TagKey: "future"}.Froze()
	err = restJson.Unmarshal(res, &data)
	if err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}
	order = data.parseOrder(symbol)
	return
}

func (e *BinanceFutureRest) FetchOpenOrders(symbol string, pageIndex, pageSize int) (orders []He_Quan.Order, err error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return
	}
	params := url.Values{}
	params.Set("symbol", market.SymbolID)
	res, err := e.Fetch(e, exchanges.Private, exchanges.GET, "/fapi/v1/openOrders", params, http.Header{})
	if err != nil {
		return
	}
	var data []Order
	restJson := jsoniter.Config{TagKey: "future"}.Froze()
	err = restJson.Unmarshal(res, &data)
	if err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}
	for _, o := range data {
		orders = append(orders, o.parseOrder(symbol))
	}
	return
}

func (e *BinanceFutureRest) FetchBalance() (balances map[string]He_Quan.Balance, err error) {
	res, err := e.Fetch(e, exchanges.Private, exchanges.GET, "/fapi/v2/balance", url.Values{}, http.Header{})
	if err != nil {
		return
	}
	var data []Balance
	restJson := jsoniter.Config{TagKey: "future"}.Froze()
	err = restJson.Unmarshal(res, &data)
	if err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}
	balances = make(map[string]He_Quan.Balance)
	for _, v := range data {
		v.Frozen = fmt.Sprintf("%f", utils.SafeParseFloat(v.Total)-utils.SafeParseFloat(v.Available))
		balances[v.Currency] = v.parseBalance()
	}
	return
}

func (e *BinanceFutureRest) FetchAccountInfo() (accountInfo He_Quan.FutureAccountInfo, err error) {
	res, err := e.Fetch(e, exchanges.Private, exchanges.GET, "/fapi/v2/account", url.Values{}, http.Header{})
	if err != nil {
		return
	}

	var data FutureAccountInfo
	if err = json.Unmarshal(res, &data); err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}

	accountInfo = data.parseAccountInfo()
	accountInfo.Positions = make(map[string]map[He_Quan.PositionType]He_Quan.FuturePositons)
	for _, position := range data.Positions {
		if math.Abs(utils.SafeParseFloat(position.Amount)) < utils.ZERO {
			continue
		}
		market, err := e.GetMarketByID(position.Symbol)
		if err != nil {
			continue
		}
		po := position.ParserFuturePosition(market.BaseID, market.Symbol)
		pos, ok := accountInfo.Positions[po.Coin]
		if !ok {
			pos = make(map[He_Quan.PositionType]He_Quan.FuturePositons)
		}
		pos[po.PositionType] = po
		accountInfo.Positions[po.Coin] = pos
	}
	return
}

func (e *BinanceFutureRest) FetchPositions(symbol string) (positions []He_Quan.FuturePositons, err error) {
	params := url.Values{}
	res, err := e.Fetch(e, exchanges.Private, exchanges.GET, "/fapi/v2/account", params, http.Header{})
	if err != nil {
		return
	}
	type datas struct {
		Positions []FuturePosition `json:"positions"`
	}
	var data datas
	if err = json.Unmarshal(res, &data); err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}
	positions = make([]He_Quan.FuturePositons, 0)
	for _, position := range data.Positions {
		market, err := e.GetMarketByID(position.Symbol)
		if err != nil {
			continue
		}
		if symbol == "" || symbol == market.Symbol {
			positions = append(positions, position.ParserFuturePosition(market.BaseID, market.Symbol))
		}

	}
	return
}

func (e *BinanceFutureRest) FetchAllPositions() (positions []He_Quan.FuturePositons, err error) {
	params := url.Values{}
	res, err := e.Fetch(e, exchanges.Private, exchanges.GET, "/fapi/v2/account", params, http.Header{})
	if err != nil {
		return
	}
	type datas struct {
		Positions []FuturePosition `json:"positions"`
	}
	var data datas
	if err = json.Unmarshal(res, &data); err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}
	positions = make([]He_Quan.FuturePositons, 0)
	for _, position := range data.Positions {
		if position.AvgPrice == "0.0" {
			continue
		}
		market, err := e.GetMarketByID(position.Symbol)
		if err != nil {
			continue
		}
		positions = append(positions, position.ParserFuturePosition(market.BaseID, market.Symbol))
	}
	return
}

func (e *BinanceFutureRest) FetchMarkPrice(symbol string) (markPrice He_Quan.MarkPrice, err error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return
	}
	params := url.Values{}
	params.Set("symbol", market.SymbolID)
	b, err := e.Fetch(e, exchanges.Public, exchanges.GET, "/fapi/v1/premiumIndex", params, http.Header{})
	if err != nil {
		return
	}
	var Mp MarkFundingRate
	err = json.Unmarshal(b, &Mp)
	if err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
	}
	markPrice = Mp.parserMarkPrice(market.Symbol)
	return
}

func (e *BinanceFutureRest) FetchFundingRate(symbol string) (fundingrate He_Quan.FundingRate, err error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return
	}
	params := url.Values{}
	params.Set("symbol", market.SymbolID)
	b, err := e.Fetch(e, exchanges.Public, exchanges.GET, "/fapi/v1/premiumIndex", params, http.Header{})
	if err != nil {
		return
	}
	var Mp MarkFundingRate
	err = json.Unmarshal(b, &Mp)
	if err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
	}
	fundingrate.Rate = Mp.LastFundingRate
	fundingrate.NextTimestamp = time.Duration(Mp.NextFundingTime)
	return
}

func (e *BinanceFutureRest) Setting(symbol string, leverage int, marginMode He_Quan.FutureMarginMode, positionMode He_Quan.FuturePositionsMode) (err error) {
	market, err := e.GetMarket(symbol)
	if err != nil {
		return
	}
	var Dual bool
	switch positionMode {
	case He_Quan.OneWay:
		Dual = false
	case He_Quan.TwoWay:
		Dual = true
	}
	var dualSide DualSidePosition
	b, err := e.Fetch(e, exchanges.Private, exchanges.GET, "/fapi/v1/positionSide/dual", url.Values{}, http.Header{})
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &dualSide)
	if err != nil {
		err = He_Quan.ExError{Code: He_Quan.ErrDataParse, Message: err.Error()}
		return
	}
	if dualSide.DaulSide != Dual {
		params := url.Values{}
		params.Set("dualSidePosition", strconv.FormatBool(Dual))
		_, err = e.Fetch(e, exchanges.Private, exchanges.POST, "/fapi/v1/positionSide/dual", params, http.Header{})
		if err != nil {
			return err
		}
	}
	symbolps, err := e.FetchPositions(symbol)
	if err != nil {
		return
	}
	if len(symbolps) > 0 && symbolps[0].MarginMode != marginMode {
		params := url.Values{}
		params.Set("symbol", market.SymbolID)
		if marginMode == He_Quan.FixedMargin {
			params.Set("marginType", "ISOLATED")
		} else {
			params.Set("marginType", "CROSSED")
		}
		_, err = e.Fetch(e, exchanges.Private, exchanges.POST, "/fapi/v1/marginType", params, http.Header{})
		if err != nil {
			return err
		}
	}
	params := url.Values{}
	params.Set("symbol", market.SymbolID)
	params.Set("leverage", strconv.Itoa(leverage))
	_, err = e.Fetch(e, exchanges.Private, exchanges.POST, "/fapi/v1/leverage", params, http.Header{})
	if err != nil {
		return err
	}
	return nil
}

func (e *BinanceFutureRest) Sign(access, method, function string, param url.Values, header http.Header) (request exchanges.Request) {
	request.Headers = header
	request.Method = method
	path := function
	if access == exchanges.Public {
		if len(param) > 0 {
			path = path + "?" + param.Encode()
		}
		request.Url = fmt.Sprintf("%s%s", e.Option.RestHost, path)
	} else {
		timeStr := fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
		param.Set("timestamp", timeStr)
		param.Set("recvWindow", "60000")
		payload := param.Encode()
		signature, err := utils.HmacSign(utils.SHA256, payload, e.Option.SecretKey, false)
		if err != nil {
			return
		}
		param.Set("signature", signature)
		if method == exchanges.GET || method == exchanges.POST || method == exchanges.DELETE {
			path = path + "?" + param.Encode()
		} else {
			request.Body = param.Encode()
		}
		request.Headers.Set("X-MBX-APIKEY", e.Option.AccessKey)
		request.Url = e.Option.RestHost + path
	}
	return request
}

func (e *BinanceFutureRest) HandleError(request exchanges.Request, response []byte) error {
	type Result struct {
		Code    int    `json:"code"`
		Message string `json:"msg"`
	}
	var result Result
	if err := json.Unmarshal(response, &result); err != nil {
		return nil
	}

	if result.Code == 0 || result.Code == 200 {
		return nil
	}
	rawErr, ok := e.errors[result.Code]
	if ok {
		if rawErr.Message == "" || strings.Contains(result.Message, rawErr.Message) {
			return He_Quan.ExError{Code: rawErr.Code, Message: result.Message}
		}
	}
	return He_Quan.ExError{Code: He_Quan.UnHandleError, Message: fmt.Sprintf("code:%v msg:%v", result.Code, result.Message)}
}
