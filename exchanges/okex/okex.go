package okex

import (
	"He_Quan"
)

type Okex struct {
	OkexRest
	OkexWs
}

func New(options He_Quan.Options) *Okex {
	instance := &Okex{}
	instance.OkexRest.Init(options)
	instance.OkexWs.Init(options)

	if len(options.Markets) == 0 {
		instance.OkexWs.Option.Markets, _ = instance.FetchMarkets()
	}

	return instance
}

