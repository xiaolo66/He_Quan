/*
@Time : 2022/12/02 1:00 下午
@Author : He_quan
@File : binance
@Software: GoLand
*/

package binance

import "github.com/xiaolo66/He_Quan"

type Binance struct {
	BinanceRest
	BinanceWs
}

func New(options He_Quan.Options) *Binance {
	instance := &Binance{}
	instance.BinanceRest.Init(options)
	instance.BinanceWs.Init(options)

	if len(options.Markets) == 0 {
		instance.BinanceWs.Option.Markets, _ = instance.FetchMarkets()
	}

	return instance
}
