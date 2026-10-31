package lagoon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/breach"
	"github.com/salsadigitalauorg/shipshape/pkg/result"
)

func TestLagoonResolveSource(t *testing.T) {
	tt := []struct {
		name         string
		pluginSource string
		resultSource string
		want         string
	}{
		{
			name:         "per-result override wins",
			pluginSource: "Shipshape",
			resultSource: "shipshape:custom",
			want:         "shipshape:custom",
		},
		{
			name:         "plugin-level used when result is empty",
			pluginSource: "shipshape:paas:docker",
			resultSource: "",
			want:         "shipshape:paas:docker",
		},
		{
			name:         "Shipshape literal when both empty",
			pluginSource: "",
			resultSource: "",
			want:         "Shipshape",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			p := &Lagoon{Source: tc.pluginSource}
			got := p.resolveSource(result.Result{Source: tc.resultSource})
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestLagoonOwnedSources(t *testing.T) {
	t.Run("plugin source plus per-result overrides, deduplicated", func(t *testing.T) {
		p := &Lagoon{Source: "Shipshape"}
		rl := &result.ResultList{
			Results: []result.Result{
				{Name: "a", Source: "shipshape:custom:a"},
				{Name: "b", Source: "shipshape:custom:a"},
				{Name: "c"},
			},
		}
		got := p.ownedSources(rl)
		assert.ElementsMatch(t, []string{"Shipshape", "shipshape:custom:a"}, got)
	})

	t.Run("passing-only results still contribute their source", func(t *testing.T) {
		p := &Lagoon{Source: "Shipshape"}
		rl := &result.ResultList{
			Results: []result.Result{
				{Name: "clean", Source: "shipshape:custom:clean", Status: result.Pass},
			},
		}
		got := p.ownedSources(rl)
		assert.ElementsMatch(t, []string{"Shipshape", "shipshape:custom:clean"}, got)
	})

	t.Run("no results still yields the plugin default", func(t *testing.T) {
		p := &Lagoon{Source: "Shipshape"}
		got := p.ownedSources(&result.ResultList{})
		assert.Equal(t, []string{"Shipshape"}, got)
	})
}

func TestLagoonOutputClearProblems(t *testing.T) {
	origGetBearerToken := getBearerToken
	defer func() { getBearerToken = origGetBearerToken }()
	getBearerToken = func() (string, error) { return "bearertoken", nil }

	t.Run("ClearProblems true issues one DELETE per distinct source", func(t *testing.T) {
		var deleteCalls []string
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodDelete {
				deleteCalls = append(deleteCalls, r.URL.Path)
			}
		})
		svr := httptest.NewServer(mux)
		defer svr.Close()

		p := &Lagoon{
			PushProblemsToInsightsRemote: true,
			InsightsRemoteEndpoint:       svr.URL + "/problems",
			InsightsRemoteBaseUrl:        svr.URL,
			Source:                       "Shipshape",
			ClearProblems:                true,
		}
		rl := &result.ResultList{
			Results: []result.Result{
				{Name: "clean"},
			},
		}

		out, err := p.Output(rl)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"/problems/Shipshape"}, deleteCalls)
		assert.Equal(t, "no breach to push to Lagoon; only deleted previous problems", string(out))
	})

	t.Run("ClearProblems true clears owned sources even with breaches", func(t *testing.T) {
		var deleteCalls []string
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodDelete {
				deleteCalls = append(deleteCalls, r.URL.Path)
				return
			}
			w.WriteHeader(200)
		})
		svr := httptest.NewServer(mux)
		defer svr.Close()

		p := &Lagoon{
			PushProblemsToInsightsRemote: true,
			InsightsRemoteEndpoint:       svr.URL + "/problems",
			InsightsRemoteBaseUrl:        svr.URL,
			Source:                       "Shipshape",
			ClearProblems:                true,
		}
		rl := &result.ResultList{
			TotalBreaches: 1,
			Results: []result.Result{
				{
					Name:     "breached",
					Breaches: []breach.Breach{&breach.ValueBreach{Value: "boom"}},
				},
			},
		}

		_, err := p.Output(rl)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"/problems/Shipshape"}, deleteCalls)
	})

	t.Run("ClearProblems false issues no DELETE", func(t *testing.T) {
		var deleteCalls []string
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodDelete {
				deleteCalls = append(deleteCalls, r.URL.Path)
				return
			}
			w.WriteHeader(200)
		})
		svr := httptest.NewServer(mux)
		defer svr.Close()

		p := &Lagoon{
			PushProblemsToInsightsRemote: true,
			InsightsRemoteEndpoint:       svr.URL + "/problems",
			InsightsRemoteBaseUrl:        svr.URL,
			Source:                       "Shipshape",
			ClearProblems:                false,
		}
		rl := &result.ResultList{}

		out, err := p.Output(rl)
		assert.NoError(t, err)
		assert.Empty(t, deleteCalls)
		assert.Equal(t, "no breach to push to Lagoon", string(out),
			"must not claim a deletion happened when ClearProblems is false")
	})
}
