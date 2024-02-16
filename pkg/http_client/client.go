package http_client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HttpClient struct {
	httpClient http.Client
}

func (h *HttpClient) RequestJson(req *http.Request, obj interface{}) (*http.Header, error) {
	res, err := h.httpClient.Do(req)
	if err != nil {
		// TODO: We can add pkg/errors here to wrap the error
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		// TODO: We can add pkg/errors here to wrap the error
		return nil, fmt.Errorf("got %d (%s) instead of 200 (OK)", res.StatusCode, res.Status)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		// TODO: We can add pkg/errors here to wrap the error
		return nil, err
	}

	err = json.Unmarshal(body, &obj)
	if err != nil {
		// TODO: We can add pkg/errors here to wrap the error
		return nil, err
	}

	return &res.Header, nil
}

func New() *HttpClient {
	httpClient := http.Client{}
	// TODO: Config http client here
	httpClient.Timeout = 5 * time.Second
	return &HttpClient{httpClient}
}
