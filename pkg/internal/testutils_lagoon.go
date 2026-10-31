package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

type MockInsightsRemoteTestState struct {
	LastCallBody     string
	LastCallHeaders  map[string]string
	LastCallEndpoint string
	LastCallStatus   string
	LastCallMethod   string
}

func MockRemoteInsightsServer(state *MockInsightsRemoteTestState) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state.LastCallEndpoint = r.RequestURI
		state.LastCallMethod = r.Method
		requestBody, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}
		state.LastCallBody = string(requestBody)
		fmt.Fprintf(w, "okay")
	}))
}
