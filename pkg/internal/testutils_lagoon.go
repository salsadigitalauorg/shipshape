package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

var MockLagoonNumCalls int
var MockLagoonRequestBodies []string

func MockLagoonServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MockLagoonNumCalls = MockLagoonNumCalls + 1
		reqBody, _ := io.ReadAll(r.Body)
		MockLagoonRequestBodies = append(MockLagoonRequestBodies, string(reqBody))
		// Response for the first request, environment id.
		if strings.Contains(string(reqBody), "environmentByKubernetesNamespaceName") {
			fmt.Fprintf(w, "{\"data\":{\"environmentByKubernetesNamespaceName\":{\"id\": 50}}}")
		} else if strings.Contains(string(reqBody), "deleteProblemsFromSource") { // Response for the deletion.
			fmt.Fprintf(w, "{\"data\":{\"deleteProblemsFromSource\":\"success\"}}")
		} else if strings.Contains(string(reqBody), "AddFactsByNameInput") { // Response for the add.
			fmt.Fprintf(w, "{}")
		} else {
			panic(string(reqBody))
		}
	}))
}

func MockLagoonReset() {
	MockLagoonNumCalls = 0
	MockLagoonRequestBodies = []string{}
}

type MockInsightsRemoteTestState struct {
	LastCallMethod   string
	LastCallBody     string
	LastCallHeaders  map[string]string
	LastCallEndpoint string
	LastCallStatus   string
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
