package syncmarkets

import (
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"github.com/trustwallet/blockatlas/pkg/blockatlas"
	"github.com/trustwallet/blockatlas/pkg/errors"
	"github.com/trustwallet/blockatlas/pkg/logger"
	"github.com/trustwallet/blockatlas/storage"
	"github.com/trustwallet/blockatlas/syncmarkets/rate"
	"github.com/trustwallet/blockatlas/syncmarkets/rate/cmc"
	"github.com/trustwallet/blockatlas/syncmarkets/rate/coingecko"
	"github.com/trustwallet/blockatlas/syncmarkets/rate/compound"
	"github.com/trustwallet/blockatlas/syncmarkets/rate/fixer"
	"math/big"
	"strings"
)

var rateProviders rate.Providers

func InitRates(storage storage.Market) {
	rateProviders = rate.Providers{
		// Add Market Quote Providers:
		0: cmc.InitRate(
			viper.GetString("market.cmc.api"),
			viper.GetString("market.cmc.api_key"),
			viper.GetString("market.cmc.map_url"),
			viper.GetString("market.rate_update_time"),
		),
		1: fixer.InitRate(
			viper.GetString("market.fixer.api"),
			viper.GetString("market.fixer.api_key"),
			viper.GetString("market.fixer.rate_update_time"),
		),
		2: compound.InitRate(
			viper.GetString("market.compound.api"),
			viper.GetString("market.rate_update_time"),
		),
		3: coingecko.InitRate(
			viper.GetString("market.coingecko.api"),
			viper.GetString("market.rate_update_time"),
		),
	}
	addRates(storage, rateProviders)
}

func addRates(storage storage.Market, rates rate.Providers) {
	c := cron.New()
	for _, r := range rates {
		scheduleTasks(storage, r, c)
	}
	c.Start()
}

func runRate(storage storage.Market, p rate.Provider) error {
	rates, err := p.FetchLatestRates()
	if err != nil {
		return errors.E(err, "FetchLatestRates")
	}
	if len(rates) > 0 {
		storage.SaveRates(rates, rateProviders)
		logger.Info("Market rates", logger.Params{"rates": len(rates), "provider": p.GetId()})
	}
	return nil
}

func GetRate(storage storage.Market, r *blockatlas.Ticker, exchangeRate float64, percentChange *big.Float) (float64, *big.Float) {
	if r.Price.Currency != blockatlas.DefaultCurrency {
		tickerRate, err := storage.GetTicker(strings.ToUpper(r.Price.Currency), "")
		if err == nil {
			exchangeRate *= tickerRate.Price.Value
			percentChange = big.NewFloat(tickerRate.Price.Change24h)
		} else {
			newRate, err := storage.GetRate(strings.ToUpper(r.Price.Currency))
			if err == nil {
				exchangeRate *= 1.0 / newRate.Rate
				percentChange = newRate.PercentChange24h
			}
		}
	}
	return exchangeRate, percentChange
}
