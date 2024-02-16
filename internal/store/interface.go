package store

type Store interface {
	Set(data map[string]float64)
	GetCurrenciesList() []string
	GetRate(currency string) (float64, bool)
	GetRates(currencies []string) []Rate
	ListRates(limit, offset int) []Rate
	Count() int
}
