package exchanges

import (
	"github.com/xiaolo66/He_Quan"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

)

type Request struct {
	Method  string
	Url     string
	Headers http.Header
	Body    string
}

const (
	Public  = "Public"
	Private = "Private"
	GET     = "GET"
	POST    = "POST"
	PUT     = "PUT"
	DELETE  = "DELETE"
)

type FetchCallBack interface {
	Sign(access, method, function string, param url.Values, header http.Header) Request
	HandleError(request Request, response []byte) error
}

type BaseExchange struct {
	Option        He_Quan.Options
	ConnectionMgr *ConnectionManager

	RwLock sync.RWMutex
}

func (b *BaseExchange) Init() {
	b.ConnectionMgr = NewConnectionManager()
	b.RwLock = sync.RWMutex{}
}

func (b *BaseExchange) GetMarketByID(symbolID string) (He_Quan.Market, error) {
	symbolID = strings.ToUpper(symbolID)
	for _, market := range b.Option.Markets {
		if market.SymbolID == symbolID {
			return market, nil
		}
	}
	return He_Quan.Market{}, errors.New(fmt.Sprintf("%v market not found", symbolID))
}

func (b *BaseExchange) GetMarket(symbol string) (He_Quan.Market, error) {
	symbol = strings.ToUpper(symbol)
	for _, market := range b.Option.Markets {
		if market.Symbol == symbol {
			return market, nil
		}
	}
	return He_Quan.Market{}, errors.New(fmt.Sprintf("%v market not found", symbol))
}

func (b *BaseExchange) Fetch(callBack FetchCallBack, access, method, function string, param url.Values, header http.Header) ([]byte, error) {
	request := callBack.Sign(access, method, function, param, header)
	client := &http.Client{}
	if b.Option.ProxyUrl != "" {
		url, _ := url.Parse(b.Option.ProxyUrl)
		client.Transport = &http.Transport{Proxy: http.ProxyURL(url)}
	}
	req, err := http.NewRequest(request.Method, request.Url, strings.NewReader(request.Body))
	if err != nil {
		return nil, He_Quan.ExError{Code: He_Quan.ErrBadRequest, Message: err.Error()}
	}
	req.Header = header

	res, err := client.Do(req)
	if err != nil {
		return nil, He_Quan.ExError{Code: He_Quan.ErrBadRequest, Message: err.Error()}
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, He_Quan.ExError{Code: He_Quan.ErrBadResponse, Message: err.Error()}
	}

	if err := callBack.HandleError(request, body); err != nil {
		return nil, err
	}
	return body, nil
}

func (b *BaseExchange) ReConnectedHandler(url string, f func()) {
	if f != nil {
		f()
	}
	//Notify subscribers of reconnection message, then clean up the channel
	//because after receiving the reconnection notification, the subscribers will resubscribe and use the new channel
	b.ConnectionMgr.PublishAfterClear(url, He_Quan.ReConnectedMessage)
}

func (b *BaseExchange) DisConnectedHandler(url string, err error, f func()) {
	// clear cache data, Prevent getting dirty data
	b.RwLock.Lock()
	defer b.RwLock.Unlock()
	if f != nil {
		f()
	}
	b.ConnectionMgr.Publish(url, He_Quan.DisConnectedMessage)
}

func (b *BaseExchange) CloseHandler(url string, f func()) {
	// clear cache data and the connection
	b.RwLock.Lock()
	defer b.RwLock.Unlock()
	if f != nil {
		f()
	}
	b.ConnectionMgr.Publish(url, He_Quan.CloseMessage)
	b.ConnectionMgr.RemoveConnection(url)
}

func (b *BaseExchange) ErrorHandler(url string, err error, f func()) {
	if f != nil {
		f()
	}
	b.ConnectionMgr.Publish(url, He_Quan.ErrorMessage(err))
}
