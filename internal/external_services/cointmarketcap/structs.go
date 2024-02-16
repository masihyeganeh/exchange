package cointmarketcap

import "time"

type Response struct {
	Data struct {
		Symbol      string    `json:"symbol"`
		Id          string    `json:"id"`
		Name        string    `json:"name"`
		Amount      int       `json:"amount"`
		LastUpdated time.Time `json:"last_updated"`
		Quote       map[string]struct {
			Price       float64   `json:"price"`
			LastUpdated time.Time `json:"last_updated"`
		} `json:"quote"`
	} `json:"data"`
	Status struct {
		Timestamp    time.Time `json:"timestamp"`
		ErrorCode    int       `json:"error_code"`
		ErrorMessage string    `json:"error_message"`
		Elapsed      int       `json:"elapsed"`
		CreditCount  int       `json:"credit_count"`
		Notice       string    `json:"notice"`
	} `json:"status"`
}
