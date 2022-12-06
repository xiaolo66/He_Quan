package factory

import (
	"github.com/xiaolo66/He_Quan"
	"github.com/xiaolo66/He_Quan/exchanges/binance"
	"github.com/xiaolo66/He_Quan/exchanges/gateio"
	"github.com/xiaolo66/He_Quan/exchanges/huobi"
	"github.com/xiaolo66/He_Quan/exchanges/okex"
)

func NewExchange(t He_Quan.ExchangeType, option He_Quan.Options) He_Quan.IExchange {
	switch t {
	case He_Quan.Binance:
		return binance.New(option)
	case He_Quan.Okex:
		return okex.New(option)
	case He_Quan.Huobi:
		return huobi.New(option)
	case He_Quan.GateIo:
		return gateio.New(option)
	}
	return nil
}

func NewFutureExchange(t He_Quan.ExchangeType, option He_Quan.Options, futureOptions He_Quan.FutureOptions) He_Quan.IFutureExchange {
	switch t {
	case He_Quan.Binance:
		return binance.NewFuture(option,futureOptions)
	}
	return nil
}
