package cointmarketcap

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/masihyeganeh/exchange/internal/external_services"
	"github.com/masihyeganeh/exchange/pkg/http_client"
)

type CoinMarketCap struct {
	httpClient     *http_client.HttpClient
	baseURL        string
	sourceCurrency string
	accessKey      string
	nextCall       time.Time
	lastUpdated    string
}

var _ external_services.ExternalApi = (*CoinMarketCap)(nil)

func (c *CoinMarketCap) MinimumInterval(interval time.Duration) time.Duration {
	if interval < MinInterval {
		return MinInterval
	}
	return interval
}

func (c *CoinMarketCap) RatesChanged(lastID string) bool {
	if c.lastUpdated != lastID {
		c.lastUpdated = lastID
		return true
	}
	return false
}

func (c *CoinMarketCap) GetRates(ctx context.Context) (updated bool, result map[string]float64, err error) {
	now := time.Now()

	if now.Before(c.nextCall) {
		// you should not call api sooner than expected
		return false, nil, nil
	}

	apiURL := c.baseURL + fmt.Sprintf("/v1/tools/price-conversion?amount=1&symbol=%s&access_key=%s", c.sourceCurrency, c.accessKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		// TODO: We can add pkg/errors here to wrap the error
		return false, nil, err
	}

	req.Header.Add("accept", "application/json")

	var data Response
	_, err = c.httpClient.RequestJson(req, &data)
	if err != nil {
		// TODO: We can add pkg/errors here to wrap the error
		return false, nil, err
	}

	if !c.RatesChanged(data.Data.LastUpdated.String()) {
		return false, nil, nil
	}

	result = make(map[string]float64, len(data.Data.Quote))

	for unit, details := range data.Data.Quote {
		unit := strings.ToLower(unit)
		result[unit] = details.Price
	}

	c.nextCall = now.Add(MinInterval)
	return true, result, nil
}

func New(httpClient *http_client.HttpClient, accessKey string) *CoinMarketCap {
	return &CoinMarketCap{
		httpClient:     httpClient,
		baseURL:        "https://pro-api.coinmarketcap.com",
		sourceCurrency: "BTC",
		accessKey:      accessKey,
	}
}
