package app

import (
	"context"
	"log"
	"time"

	"github.com/masihyeganeh/exchange/internal/external_services"
	"github.com/masihyeganeh/exchange/internal/store"
)

type App struct {
	store  store.Store
	api    external_services.ExternalApi
	cancel func()
	Done   chan error
}

// New creates a new instance of the App
func New(store store.Store, api external_services.ExternalApi) *App {
	return &App{store: store, api: api}
}

func (a *App) Update(ctx context.Context) {
	updated, result, err := a.api.GetRates(ctx)
	if err != nil {
		log.Println(err.Error())
		return
	}

	if !updated {
		log.Println("rates are still the same")
		return
	}

	a.store.Set(result)
	log.Println("rates are updated")
}

func (a *App) StartWorker(ctx context.Context, interval time.Duration) {
	actualInterval := a.api.MinimumInterval(interval)

	a.Update(ctx)

	ticker := time.NewTicker(actualInterval)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				go a.Update(ctx)
			}
		}
	}()
}
