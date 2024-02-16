// Package exchange provides api for exchanges
package exchange

import (
	"context"

	"github.com/masihyeganeh/exchange/api/exchange"
	"github.com/masihyeganeh/exchange/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// API represents implementation of Exchange Provisioning server API
type API struct {
	exchange.UnimplementedExchangeServer
	store store.Store
}

// New creates a new instance of server
func New(store store.Store) exchange.ExchangeServer {
	return &API{
		store: store,
	}
}

// Convert - Convert the amount of crypto to the given currency.
func (a *API) Convert(_ context.Context, req *exchange.ConvertRequest) (*exchange.ConvertResponse, error) {
	rate, exists := a.store.GetRate(req.Currency)
	if !exists {
		return nil, status.Errorf(codes.Internal, "unknown currency: %s", req.Currency)
	}

	return &exchange.ConvertResponse{
		Value:    req.Amount * float32(rate),
		Currency: req.Currency,
	}, nil
}

// BatchConvert - Convert the amount of crypto to the given currencies.
func (a *API) BatchConvert(_ context.Context, req *exchange.BatchConvertRequest) (*exchange.BatchConvertResponse, error) {
	rates := a.store.GetRates(req.Currencies)
	if len(rates) == 0 {
		return nil, status.Error(codes.Internal, "none of the currencies was valid")
	}

	result := &exchange.BatchConvertResponse{
		List: make([]*exchange.ConvertResponse, 0),
	}
	for _, rate := range rates {
		result.List = append(result.List, &exchange.ConvertResponse{
			Value:    req.Amount * float32(rate.Rate),
			Currency: rate.Currency,
		})
	}

	return result, nil
}

// ListRates - listing of rates with pagination.
func (a *API) ListRates(_ context.Context, req *exchange.ListRatesRequest) (*exchange.ListRatesResponse, error) {
	if req.PageSize == int32(0) {
		req.PageSize = int32(5)
	}

	rates := a.store.ListRates(int(req.PageSize), int(req.PageId))
	if len(rates) == 0 {
		return nil, status.Error(codes.Internal, "we have no rates now")
	}

	result := &exchange.ListRatesResponse{
		Items: make([]*exchange.Rate, 0),
	}
	for _, rate := range rates {
		result.Items = append(result.Items, &exchange.Rate{
			ConversionRate: float32(rate.Rate),
			Currency:       rate.Currency,
		})
	}

	result.NextPage = int32(0)
	if req.PageSize*(req.PageId+1) < int32(a.store.Count()) {
		result.NextPage = req.PageId + 1
	}

	return result, nil
}
