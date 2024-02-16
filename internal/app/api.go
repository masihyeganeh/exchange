package app

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	root "github.com/masihyeganeh/exchange"
	"github.com/masihyeganeh/exchange/api/exchange"
	exchangeApi "github.com/masihyeganeh/exchange/internal/app/exchange"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

func (a *App) StartServer(ctx context.Context, bind string) error {
	ctx, a.cancel = context.WithCancel(ctx)
	gs := grpc.NewServer()

	exchange.RegisterExchangeServer(gs, exchangeApi.New(a.store))

	// sets up the client http interface variable
	// we need to start the gRPC service first, as it is used by the
	// grpc-gateway
	var clientHTTPHandler http.Handler

	// switch between gRPC and "plain" http handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			gs.ServeHTTP(w, r)
		} else {
			if clientHTTPHandler == nil {
				w.WriteHeader(http.StatusNotImplemented)
				return
			}

			//if a.cfg.CORSAllowOrigin != "" {
			//	w.Header().Set("Access-Control-Allow-Origin", a.cfg.CORSAllowOrigin)
			//	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			//	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Grpc-Metadata-Authorization, Grpc-Metadata-X-OTP")
			//
			//	if r.Method == "OPTIONS" {
			//		return
			//	}
			//}

			clientHTTPHandler.ServeHTTP(w, r)
		}
	})

	// start the API server
	go func() {
		log.Println("starting api server")

		server := &http.Server{
			Addr:              bind,
			ReadHeaderTimeout: 3 * time.Second,
			Handler:           h2c.NewHandler(handler, &http2.Server{}),
		}
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal(err.Error())
		}

		a.Done <- err
		close(a.Done)
	}()

	// give the http server some time to start
	time.Sleep(time.Millisecond * 100)

	// sets up the HTTP handler
	clientHTTPHandler, err := a.setupHTTPAPI(bind)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		gs.GracefulStop()
	}()
	return nil
}

func (a *App) setupHTTPAPI(bind string) (http.Handler, error) {
	r := mux.NewRouter()

	// setup json api handler
	jsonHandler, err := a.getJSONGateway(context.Background(), bind)
	if err != nil {
		return nil, err
	}

	log.Println("registering rest api handler and documentation endpoint")
	staticFs, err := fs.Sub(root.StaticFiles, "static")
	if err != nil {
		return nil, err
	}

	swaggerIndex, swaggerIndexErr := fs.ReadFile(root.StaticFiles, "static/swagger/index.html")
	if swaggerIndexErr != nil {
		return nil, swaggerIndexErr
	}

	r.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		if swaggerIndexErr != nil {
			log.Println("get swagger template error")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(swaggerIndex)
	}).Methods("get")
	r.PathPrefix("/api").Handler(jsonHandler)

	r.Path("/").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("location", "/api")
		writer.WriteHeader(http.StatusFound)
	})
	r.PathPrefix("/").Handler(http.FileServer(http.FS(staticFs))).Methods("get")

	return r, nil
}

func (a *App) getJSONGateway(ctx context.Context, bind string) (http.Handler, error) {
	// dial options for the grpc-gateway
	var grpcDialOpts []grpc.DialOption

	grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	bindParts := strings.SplitN(bind, ":", 2)
	if len(bindParts) != 2 {
		log.Fatal("get port from bind failed")
	}
	apiEndpoint := fmt.Sprintf("localhost:%s", bindParts[1])

	serveMux := runtime.NewServeMux(runtime.WithMarshalerOption(
		runtime.MIMEWildcard,
		&runtime.JSONPb{
			EnumsAsInts:  false,
			EmitDefaults: true,
		},
	))

	err := exchange.RegisterExchangeHandlerFromEndpoint(ctx, serveMux, apiEndpoint, grpcDialOpts)
	if err != nil {
		return nil, err
	}

	return serveMux, nil
}

// StopServer stops the gRPC service
func (a *App) StopServer() error {
	a.cancel()
	return <-a.Done
}
