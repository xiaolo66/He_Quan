package coinbase

import "He_Quan"

type CoinBase struct {
	CoinBaseRest
	CoinBaseWs
}

func New(options He_Quan.Options) *CoinBase {
	instance := &CoinBase{}
	instance.CoinBaseRest.Init(options)
	instance.CoinBaseWs.Init(options)

	if len(options.Markets) == 0 {
		instance.CoinBaseWs.Option.Markets, _ = instance.FetchMarkets()
	}
	return instance
}