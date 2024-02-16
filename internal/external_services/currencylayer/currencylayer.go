package currencylayer

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/masihyeganeh/exchange/internal/external_services"
	"github.com/masihyeganeh/exchange/pkg/http_client"
)

type CurrencyLayer struct {
	httpClient     *http_client.HttpClient
	baseURL        string
	sourceCurrency string
	accessKey      string
	nextCall       time.Time
	etag           string
}

var _ external_services.ExternalApi = (*CurrencyLayer)(nil)

func (c *CurrencyLayer) MinimumInterval(interval time.Duration) time.Duration {
	if interval < MinInterval {
		return MinInterval
	}
	return interval
}

func (c *CurrencyLayer) RatesChanged(lastID string) bool {
	if c.etag != lastID {
		c.etag = lastID
		return true
	}
	return false
}

func (c *CurrencyLayer) GetRates(ctx context.Context) (updated bool, result map[string]float64, err error) {
	now := time.Now()

	if now.Before(c.nextCall) {
		// you should not call api sooner than expected
		return false, nil, nil
	}

	apiURL := c.baseURL + fmt.Sprintf("/live?source=%s&access_key=%s", c.sourceCurrency, c.accessKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		// TODO: We can add pkg/errors here to wrap the error
		return false, nil, err
	}

	req.Header.Add("accept", "application/json")

	var data Response
	header, err := c.httpClient.RequestJson(req, &data)
	if err != nil {
		// TODO: We can add pkg/errors here to wrap the error
		return false, nil, err
	}

	if !c.RatesChanged(header.Get("etag")) {
		return false, nil, nil
	}

	result = make(map[string]float64, len(data.Quotes))

	for unit, rate := range data.Quotes {
		unit := strings.ToLower(strings.Replace(unit, c.sourceCurrency, "", 1))
		result[unit] = rate
	}

	c.nextCall = now.Add(MinInterval)
	return true, result, nil
}

func New(httpClient *http_client.HttpClient, accessKey string) *CurrencyLayer {
	return &CurrencyLayer{
		httpClient:     httpClient,
		baseURL:        "https://api.currencylayer.com",
		sourceCurrency: "BTC",
		accessKey:      accessKey,
	}
}
