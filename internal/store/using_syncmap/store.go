package using_syncmap

import (
	"slices"
	"sort"

	base "github.com/masihyeganeh/exchange/internal/store"
	"golang.org/x/sync/syncmap"
)

type Store struct {
	data       syncmap.Map
	currencies []string
}

func (s *Store) Set(data map[string]float64) {
	currencies := make([]string, 0, len(data))
	for currency, rate := range data {
		s.data.Store(currency, rate)
		currencies = append(currencies, currency)
	}

	sort.Strings(currencies)

	s.currencies = currencies
}

func (s *Store) GetCurrenciesList() []string {
	return slices.Clone(s.currencies)
}

func (s *Store) GetRate(currency string) (float64, bool) {

	value, exists := s.data.Load(currency)
	return value.(float64), exists
}

func (s *Store) GetRates(currencies []string) []base.Rate {
	result := make([]base.Rate, 0, len(currencies))

	for _, currency := range currencies {
		if value, exists := s.data.Load(currency); exists {
			result = append(result, base.Rate{
				Currency: currency,
				Rate:     value.(float64),
			})
		}
	}
	return result
}

func (s *Store) ListRates(limit, offset int) []base.Rate {
	result := make([]base.Rate, 0, limit)

	if limit < 0 || offset < 0 {
		return result
	}

	if len(s.currencies) < limit+offset {
		return result
	}

	currencies := s.currencies[offset:limit]
	for _, currency := range currencies {
		rate, exists := s.data.Load(currency)
		if exists {
			result = append(result, base.Rate{
				Currency: currency,
				Rate:     rate.(float64),
			})
		}
	}

	return result
}

func (s *Store) Count() int {
	return len(s.currencies)
}

var _ base.Store = (*Store)(nil)

func New() *Store {
	return &Store{}
}
