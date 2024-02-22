package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/masihyeganeh/exchange/internal/app"
	"github.com/masihyeganeh/exchange/internal/external_services/coingecko"
	"github.com/masihyeganeh/exchange/internal/store/using_mutex"
	//"github.com/masihyeganeh/exchange/internal/store/using_syncmap"
	//"github.com/masihyeganeh/exchange/internal/store/using_atomic"
	"github.com/masihyeganeh/exchange/pkg/http_client"
)

func main() {
	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGKILL)

	httpClient := http_client.New()

	// You can switch between external APIs by uncommenting others (in a more serious
	// implementation, one can be fallback of the other or all of them can be called with round-robin)
	externalApi := coingecko.New(httpClient)
	//externalApi := cointmarketcap.New(httpClient, "access key")
	//externalApi := currencylayer.New(httpClient, "access key")

	// You can switch between stores by uncommenting others
	store := using_mutex.New()
	//store := using_syncmap.New()
	//store := using_atomic.New()

	application := app.New(store, externalApi)

	application.StartWorker(ctx, 1*time.Minute)

	err := application.StartServer(ctx, ":8123")
	if err != nil {
		log.Fatal(err.Error())
	}

	select {
	case err := <-application.Done:
		log.Println("application has exited")
		if err != nil {
			log.Println(err.Error())
		}
	case _ = <-ctx.Done():
		log.Println("stop signal received, exiting")
		stop()
		err := application.StopServer()
		if err != nil {
			log.Println(err.Error())
		}
	}
}
