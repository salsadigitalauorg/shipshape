package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/salsadigitalauorg/shipshape/pkg/data"
	"github.com/salsadigitalauorg/shipshape/pkg/fact"
	. "github.com/salsadigitalauorg/shipshape/pkg/fact/http"
)

func TestFetchInit(t *testing.T) {
	assert := assert.New(t)

	factPlugin := fact.Manager().GetFactories()["http:fetch"]("testFetch")
	assert.NotNil(factPlugin)
	fetchFacter, ok := factPlugin.(*Fetch)
	assert.True(ok)
	assert.Equal("testFetch", fetchFacter.GetId())
}

func TestFetchPluginName(t *testing.T) {
	assert.Equal(t, "http:fetch", NewFetch("testFetch").GetName())
}

func TestFetchNoUrl(t *testing.T) {
	f := NewFetch("testFetch")
	f.Collect()
	assert.Equal(t, []error{ErrNoUrl}, f.GetErrors())
}

func TestFetchInvalidUrl(t *testing.T) {
	f := NewFetch("testFetch")
	f.Url = "not-a-url"
	f.Collect()
	assert.Equal(t, []error{ErrInvalidUrl}, f.GetErrors())
}

func TestFetchInsecureSchemeRejectedByDefault(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("should not be fetched"))
	}))
	defer svr.Close()

	f := NewFetch("testFetch")
	f.Url = svr.URL // httptest.NewServer is plain http://
	f.Collect()
	assert.Equal(t, []error{ErrSchemeNotAllowed}, f.GetErrors())
}

func TestFetchInsecureSchemeAllowedWhenOptedIn(t *testing.T) {
	expected := "template content"
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(expected))
	}))
	defer svr.Close()

	f := NewFetch("testFetch")
	f.Url = svr.URL
	f.AllowInsecure = true
	f.Collect()
	assert.Empty(t, f.GetErrors())
	assert.Equal(t, data.FormatRaw, f.GetFormat())
	assert.Equal(t, []byte(expected), f.GetData())
}

func TestFetchNotFoundHintsAtLocalClone(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404: Not Found"))
	}))
	defer svr.Close()

	f := NewFetch("testFetch")
	f.Url = svr.URL
	f.AllowInsecure = true
	f.Collect()

	assert.Len(t, f.GetErrors(), 1)
	assert.ErrorContains(t, f.GetErrors()[0], "file:read")
	assert.ErrorContains(t, f.GetErrors()[0], "404")
}
