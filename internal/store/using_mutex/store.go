package using_mutex

import (
	"slices"
	"sort"
	"sync"

	base "github.com/masihyeganeh/exchange/internal/store"
)

type Store struct {
	data       map[string]float64
	currencies []string
	lock       sync.RWMutex
}

func (s *Store) Set(data map[string]float64) {
	currencies := make([]string, 0, len(data))
	for currency := range data {
		currencies = append(currencies, currency)
	}

	sort.Strings(currencies)

	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = data
	s.currencies = currencies
}

func (s *Store) GetCurrenciesList() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return slices.Clone(s.currencies)
}

func (s *Store) GetRate(currency string) (float64, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	value, exists := s.data[currency]
	return value, exists
}

func (s *Store) GetRates(currencies []string) []base.Rate {
	s.lock.RLock()
	defer s.lock.RUnlock()

	result := make([]base.Rate, 0, len(currencies))

	for _, currency := range currencies {
		if value, exists := s.data[currency]; exists {
			result = append(result, base.Rate{
				Currency: currency,
				Rate:     value,
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

	s.lock.RLock()
	defer s.lock.RUnlock()

	if len(s.currencies) < limit+offset {
		return result
	}

	currencies := s.currencies[offset:limit]
	for _, currency := range currencies {
		rate, exists := s.data[currency]
		if exists {
			result = append(result, base.Rate{
				Currency: currency,
				Rate:     rate,
			})
		}
	}

	return result
}

func (s *Store) Count() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.currencies)
}

var _ base.Store = (*Store)(nil)

func New() *Store {
	return &Store{}
}
