package lagoon_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/salsadigitalauorg/shipshape/pkg/internal"
	"github.com/salsadigitalauorg/shipshape/pkg/lagoon"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMustHaveEnvVars(t *testing.T) {
	assert := assert.New(t)

	origOutput := logrus.StandardLogger().Out
	defer func() {
		logrus.StandardLogger().ExitFunc = nil
		logrus.SetOutput(origOutput)
	}()

	var buf bytes.Buffer
	logrus.SetOutput(&buf)

	var fatal bool
	logrus.StandardLogger().ExitFunc = func(retCode int) {
		fatal = (retCode == 1)
	}

	t.Run("varsSet", func(t *testing.T) {
		os.Setenv("LAGOON_PROJECT", "foo")
		os.Setenv("LAGOON_ENVIRONMENT", "bar")
		defer func() {
			os.Unsetenv("LAGOON_PROJECT")
			os.Unsetenv("LAGOON_ENVIRONMENT")
		}()

		o := lagoon.Lagoon{Source: "Shipshape"}
		o.EnvironmentOverrides()
		o.MustHaveEnvVars()
		assert.False(fatal, "Got this fatal: "+buf.String())
	})

	t.Run("varsUnset", func(t *testing.T) {
		o := lagoon.Lagoon{}
		o.MustHaveEnvVars()
		assert.True(fatal)
		assert.Contains(buf.String(), "project & environment name required; please ensure both LAGOON_PROJECT & LAGOON_ENVIRONMENT are set")
	})
}

func Test_DeleteProblemsInsightsRemote(t *testing.T) {
	t.Run("single source", func(t *testing.T) {
		state := internal.MockInsightsRemoteTestState{}
		svr := internal.MockRemoteInsightsServer(&state)
		defer svr.Close()

		err := lagoon.DeleteProblems([]string{"Shipshape"}, svr.URL, "bearertoken")
		assert.NoError(t, err)
		assert.Equal(t, http.MethodDelete, state.LastCallMethod)
		assert.Equal(t, "/problems/Shipshape", state.LastCallEndpoint)
	})

	t.Run("source containing colons round-trips as a single path segment", func(t *testing.T) {
		// url.PathEscape does not percent-encode ':' - it is a valid pchar
		// per RFC 3986 - so the literal colon is expected here. The point of
		// using PathEscape is correctness for characters that DO need
		// escaping (e.g. '/', '?', '#', spaces), not colons specifically.
		state := internal.MockInsightsRemoteTestState{}
		svr := internal.MockRemoteInsightsServer(&state)
		defer svr.Close()

		err := lagoon.DeleteProblems([]string{"shipshape:paas:docker"}, svr.URL, "bearertoken")
		assert.NoError(t, err)
		assert.Equal(t, http.MethodDelete, state.LastCallMethod)
		assert.Equal(t, "/problems/shipshape:paas:docker", state.LastCallEndpoint)
	})

	t.Run("one request per distinct source, duplicates skipped", func(t *testing.T) {
		var calls []string
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			calls = append(calls, r.URL.Path)
		})
		svr := httptest.NewServer(mux)
		defer svr.Close()

		err := lagoon.DeleteProblems(
			[]string{"a", "b", "a", "b", "c"}, svr.URL, "bearertoken")
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"/problems/a", "/problems/b", "/problems/c"}, calls)
	})

	t.Run("base url with trailing slash", func(t *testing.T) {
		state := internal.MockInsightsRemoteTestState{}
		svr := internal.MockRemoteInsightsServer(&state)
		defer svr.Close()

		err := lagoon.DeleteProblems([]string{"Shipshape"}, svr.URL+"/", "bearertoken")
		assert.NoError(t, err)
		assert.Equal(t, "/problems/Shipshape", state.LastCallEndpoint)
	})
}

func Test_GetBearerTokenFromDisk(t *testing.T) {
	type args struct {
		tokenLocation string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "Successfully loads token from disk",
			args: args{
				tokenLocation: "./testdata/insightsbearertoken",
			},
			want:    "bearertokentext",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lagoon.GetBearerTokenFromDisk(tt.args.tokenLocation)
			if !tt.wantErr(t, err, fmt.Sprintf("getBearerTokenFromDisk(%v)", tt.args.tokenLocation)) {
				return
			}
			assert.Equalf(t, tt.want, got, "getBearerTokenFromDisk(%v)", tt.args.tokenLocation)
		})
	}
}

func Test_ProblemsToInsightsRemote(t *testing.T) {
	type args struct {
		problems    []lagoon.Problem
		bearerToken string
	}
	tests := []struct {
		name          string
		args          args
		wantErr       assert.ErrorAssertionFunc
		testBodyEqual bool
	}{
		{
			name: "Successful post",
			args: args{
				problems: []lagoon.Problem{
					{
						EnvironmentId: 1,
						Identifier:    "problem1",
						Severity:      "HIGH",
						Service:       "serviceName",
					},
				},
				bearerToken: "bearertoken",
			},
			wantErr:       assert.NoError,
			testBodyEqual: true,
		},
	}
	for _, tt := range tests {

		mockServerData := internal.MockInsightsRemoteTestState{}

		serv := internal.MockRemoteInsightsServer(&mockServerData)

		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, lagoon.ProblemsToInsightsRemote(tt.args.problems, serv.URL, tt.args.bearerToken), fmt.Sprintf("problemsToInsightsRemote(%v, %v)", tt.args.problems, tt.args.bearerToken))
			if tt.testBodyEqual == true { //let's check the data we sent through seems correct
				bodyString := mockServerData.LastCallBody
				var problems []lagoon.Problem
				err := json.Unmarshal([]byte(bodyString), &problems)
				assert.NoError(t, err)
				assert.Equalf(t, tt.args.problems, problems, "Unmarshalled Body not Equal")
			}
		})
	}
}
