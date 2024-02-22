package using_atomic

import (
	"slices"
	"sort"
	"sync"
	"sync/atomic"

	base "github.com/masihyeganeh/exchange/internal/store"
)

type Store struct {
	data                map[string]float64
	secondaryData       map[string]float64
	currencies          []string
	secondaryCurrencies []string
	readSecondary       atomic.Bool
	lock                sync.RWMutex
}

// Set fills the map with new data. It is assumed that we only
// have one writer that doesn't call this method concurrently
func (s *Store) Set(data map[string]float64) {
	currencies := make([]string, 0, len(data))
	for currency := range data {
		currencies = append(currencies, currency)
	}

	sort.Strings(currencies)

	if s.readSecondary.Load() {
		s.data = data
		s.currencies = currencies
		s.lock.Lock()
		s.readSecondary.Store(false)
		s.lock.Unlock()
	} else {
		s.secondaryData = data
		s.secondaryCurrencies = currencies
		s.lock.Lock()
		s.readSecondary.Store(true)
		s.lock.Unlock()
	}
}

func (s *Store) GetCurrenciesList() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.readSecondary.Load() {
		return slices.Clone(s.secondaryCurrencies)
	} else {
		return slices.Clone(s.currencies)
	}
}

func (s *Store) GetRate(currency string) (float64, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.readSecondary.Load() {
		value, exists := s.secondaryData[currency]
		return value, exists
	} else {
		value, exists := s.data[currency]
		return value, exists
	}
}

func (s *Store) GetRates(currencies []string) []base.Rate {
	s.lock.RLock()
	defer s.lock.RUnlock()

	result := make([]base.Rate, 0, len(currencies))

	if s.readSecondary.Load() {
		for _, currency := range currencies {
			if value, exists := s.secondaryData[currency]; exists {
				result = append(result, base.Rate{
					Currency: currency,
					Rate:     value,
				})
			}
		}
	} else {
		for _, currency := range currencies {
			if value, exists := s.data[currency]; exists {
				result = append(result, base.Rate{
					Currency: currency,
					Rate:     value,
				})
			}
		}
	}
	return result
}

func (s *Store) ListRates(limit, offset int) []base.Rate {
	result := make([]base.Rate, 0, limit)

	if limit < 0 || offset < 0 {
		return result
	}

	if s.Count() < limit+offset {
		return result
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.readSecondary.Load() {
		currencies := s.secondaryCurrencies[offset:limit]
		for _, currency := range currencies {
			rate, exists := s.secondaryData[currency]
			if exists {
				result = append(result, base.Rate{
					Currency: currency,
					Rate:     rate,
				})
			}
		}
	} else {
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
	}

	return result
}

func (s *Store) Count() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if s.readSecondary.Load() {
		return len(s.secondaryCurrencies)
	} else {
		return len(s.currencies)
	}
}

var _ base.Store = (*Store)(nil)

func New() *Store {
	return &Store{}
}
