package coingecko

import (
	"context"
	"net/http"
	"time"

	"github.com/masihyeganeh/exchange/internal/external_services"
	"github.com/masihyeganeh/exchange/pkg/http_client"
)

var _ external_services.ExternalApi = (*CoinGecko)(nil)

type CoinGecko struct {
	httpClient *http_client.HttpClient
	baseURL    string
	nextCall   time.Time
}

func (c *CoinGecko) MinimumInterval(interval time.Duration) time.Duration {
	if interval < MinInterval {
		return MinInterval
	}
	return interval
}

func (c *CoinGecko) RatesChanged(_ string) bool {
	return true
}

func (c *CoinGecko) GetRates(ctx context.Context) (updated bool, result map[string]float64, err error) {
	now := time.Now()

	if now.Before(c.nextCall) {
		// you should not call api sooner than expected
		return false, nil, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/exchange_rates", nil)
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

	result = make(map[string]float64, len(data.Rates))

	for unit, details := range data.Rates {
		// We explicitly need conversion rates to fiat
		if details.Type != "fiat" {
			continue
		}
		result[unit] = details.Value
	}

	c.nextCall = now.Add(MinInterval)
	return true, result, nil
}

func New(httpClient *http_client.HttpClient) *CoinGecko {
	return &CoinGecko{
		httpClient: httpClient,
		baseURL:    "https://api.coingecko.com/api/v3",
	}
}
