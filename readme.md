# Exchange

Converts crypto to fiat currencies.  
For now, we assume that the source crypto to exchange is `BTC`.

## Folder structure

- `/api`: contains .proto files as well as protobuf files and swagger files generated from them.
- `/cmd`: contains main file that acts as a dependency manager and starts the app.
- `/internal`: contains the actual code of the system that should not be exported.
- - `/internal/app`: contains the `app` that initializes the `api` and in sub-folders, there are the actual
implementation of the APIs.
- - `/internal/external_services`: contains different implementations of the API call to get the convert rate
from external services. All of them are implementation based on the `ExternalApi` interface.
- - - `/internal/store`: contains different implementations of the storage layer.
- `/pkg`: contains packages that are OK to export.
- - `/pkg/http_client`: is a simple wrapper on the `http.Client` of Go, that simplifies API calls and unmarshalling.
- `/static`: contains generated swagger file with the static files needed to serve it.

## External API call design

The interface for external API calls has 3 methods:
- `MinimumInterval`: Each service that we are using has its own time limit to call the API. this method helps with
finding the best interval to call the API again.
- `RatesChanged`: Some of the services provide ways to check if the result from them is the same or not. This method
helps to find out if it is actually changed, so we don't need to change or data. It's just a performance optimization.
- `GetRates`: The actual method to get the rates.

## Storage layer implementations

I have 3 implementations for the storage layer:
1. `using_mutex`: Simple `map` with `RWMutex` which is usually more than enough even in high traffics.
2. `using_syncmap`: Uses `syncmap` which is a concurrent map implementation but should be used when we made sure
that the bottleneck is the map.
3. `using_atomic`: Which has two maps to that writes to one and reads from the other and when the writer can acquire the
lock, it changes an `atomic` flag that redirects reader to the alternative map and updates current map.
