package external_services

import (
	"context"
	"time"
)

type ExternalApi interface {
	MinimumInterval(interval time.Duration) time.Duration
	RatesChanged(lastID string) bool
	GetRates(ctx context.Context) (updated bool, result map[string]float64, err error)
}
