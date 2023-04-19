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
		} else if strings.Contains(string(reqBody), "deleteFactsFromSource") { // Response for the deletion.
			fmt.Fprintf(w, "{\"data\":{\"deleteFactsFromSource\":\"success\"}}")
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
