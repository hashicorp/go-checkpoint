package checkpoint

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/madlambda/spells/assert"
)

func TestCheck(t *testing.T) {
	expected := &CheckResponse{
		Product:             "test",
		CurrentVersion:      "1.0",
		CurrentReleaseDate:  0,
		CurrentDownloadURL:  "http://www.hashicorp.com",
		CurrentChangelogURL: "http://www.hashicorp.com",
		ProjectWebsite:      "http://www.hashicorp.com",
		Outdated:            false,
		Alerts:              []*CheckAlert{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		assert.EqualStrings(t, "/v1/check/test", r.URL.Path)
		assert.EqualStrings(t, "1.0", params.Get("version"))
		assert.EqualStrings(t, runtime.GOARCH, params.Get("arch"))
		assert.EqualStrings(t, runtime.GOOS, params.Get("os"))
		assert.EqualStrings(t, "", params.Get("signature"))

		retdata, err := json.Marshal(expected)
		assert.NoError(t, err)
		w.Write(retdata)
	}))

	endpoint, err := url.Parse(server.URL)
	assert.NoError(t, err)
	actual, err := CheckAt(*endpoint, &CheckParams{
		Product: "test",
		Version: "1.0",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %#v, got: %#v", expected, actual)
	}
}

func TestCheck_timeout(t *testing.T) {
	os.Setenv("CHECKPOINT_TIMEOUT", "50")
	defer os.Setenv("CHECKPOINT_TIMEOUT", "")

	expected := "Client.Timeout exceeded while awaiting headers"

	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		assert.EqualStrings(t, "/v1/check/test", r.URL.Path)
		assert.EqualStrings(t, "1.0", params.Get("version"))
		assert.EqualStrings(t, runtime.GOARCH, params.Get("arch"))
		assert.EqualStrings(t, runtime.GOOS, params.Get("os"))
		assert.EqualStrings(t, "", params.Get("signature"))

		time.Sleep(1 * time.Minute)
	}))

	endpoint, err := url.Parse(slowServer.URL)
	assert.NoError(t, err)

	_, err = CheckAt(*endpoint, &CheckParams{
		Product: "test",
		Version: "1.0",
	})

	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected a timeout error, got: %v", err)
	}
}

func TestCheck_disabled(t *testing.T) {
	os.Setenv("CHECKPOINT_DISABLE", "1")
	defer os.Setenv("CHECKPOINT_DISABLE", "")

	expected := &CheckResponse{}

	actual, err := Check(&CheckParams{
		Product: "test",
		Version: "1.0",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %#v, got: %#v", expected, actual)
	}
}

func TestCheck_cache(t *testing.T) {
	dir, err := ioutil.TempDir("", "checkpoint")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := &CheckResponse{
		Product:             "test",
		CurrentVersion:      "1.0",
		CurrentReleaseDate:  0,
		CurrentDownloadURL:  "http://www.hashicorp.com",
		CurrentChangelogURL: "http://www.hashicorp.com",
		ProjectWebsite:      "http://www.hashicorp.com",
		Outdated:            false,
		Alerts:              []*CheckAlert{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		assert.EqualStrings(t, "/v1/check/test", r.URL.Path)
		assert.EqualStrings(t, "1.0", params.Get("version"))
		assert.EqualStrings(t, runtime.GOARCH, params.Get("arch"))
		assert.EqualStrings(t, runtime.GOOS, params.Get("os"))
		assert.EqualStrings(t, "", params.Get("signature"))

		retdata, err := json.Marshal(expected)
		assert.NoError(t, err)
		w.Write(retdata)
	}))

	endpoint, err := url.Parse(server.URL)
	assert.NoError(t, err)

	var actual *CheckResponse
	for i := 0; i < 5; i++ {
		var err error
		actual, err = CheckAt(*endpoint, &CheckParams{
			Product:   "test",
			Version:   "1.0",
			CacheFile: filepath.Join(dir, "cache"),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %#v, got: %#v", expected, actual)
	}
}

func TestCheck_cacheNested(t *testing.T) {
	dir, err := ioutil.TempDir("", "checkpoint")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := &CheckResponse{
		Product:             "test",
		CurrentVersion:      "1.0",
		CurrentReleaseDate:  0,
		CurrentDownloadURL:  "http://www.hashicorp.com",
		CurrentChangelogURL: "http://www.hashicorp.com",
		ProjectWebsite:      "http://www.hashicorp.com",
		Outdated:            false,
		Alerts:              []*CheckAlert{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		assert.EqualStrings(t, "/v1/check/test", r.URL.Path)
		assert.EqualStrings(t, "1.0", params.Get("version"))
		assert.EqualStrings(t, runtime.GOARCH, params.Get("arch"))
		assert.EqualStrings(t, runtime.GOOS, params.Get("os"))
		assert.EqualStrings(t, "", params.Get("signature"))

		retdata, err := json.Marshal(expected)
		assert.NoError(t, err)
		w.Write(retdata)
	}))

	endpoint, err := url.Parse(server.URL)
	assert.NoError(t, err)

	var actual *CheckResponse
	for i := 0; i < 5; i++ {
		var err error
		actual, err = CheckAt(*endpoint, &CheckParams{
			Product:   "test",
			Version:   "1.0",
			CacheFile: filepath.Join(dir, "nested", "cache"),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %#v, got: %#v", expected, actual)
	}
}

func TestCheckInterval(t *testing.T) {
	expected := &CheckResponse{
		Product:             "test",
		CurrentVersion:      "1.0",
		CurrentReleaseDate:  0,
		CurrentDownloadURL:  "http://www.hashicorp.com",
		CurrentChangelogURL: "http://www.hashicorp.com",
		ProjectWebsite:      "http://www.hashicorp.com",
		Outdated:            false,
		Alerts:              []*CheckAlert{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		assert.EqualStrings(t, "/v1/check/test", r.URL.Path)
		assert.EqualStrings(t, "1.0", params.Get("version"))
		assert.EqualStrings(t, runtime.GOARCH, params.Get("arch"))
		assert.EqualStrings(t, runtime.GOOS, params.Get("os"))
		assert.EqualStrings(t, "", params.Get("signature"))

		retdata, err := json.Marshal(expected)
		assert.NoError(t, err)
		w.Write(retdata)
	}))

	endpoint, err := url.Parse(server.URL)
	assert.NoError(t, err)

	params := &CheckParams{
		Product: "test",
		Version: "1.0",
	}

	calledCh := make(chan struct{})
	checkFn := func(actual *CheckResponse, err error) {
		defer close(calledCh)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("expected %#v, got: %#v", expected, actual)
		}
	}

	doneCh := CheckIntervalAt(*endpoint, params, 500*time.Millisecond, checkFn)
	defer close(doneCh)

	select {
	case <-calledCh:
	case <-time.After(1250 * time.Millisecond):
		t.Fatalf("timeout")
	}
}

func TestCheckInterval_disabled(t *testing.T) {
	os.Setenv("CHECKPOINT_DISABLE", "1")
	defer os.Setenv("CHECKPOINT_DISABLE", "")

	params := &CheckParams{
		Product: "test",
		Version: "1.0",
	}

	calledCh := make(chan struct{})
	checkFn := func(actual *CheckResponse, err error) {
		defer close(calledCh)
	}

	doneCh := CheckInterval(params, 500*time.Millisecond, checkFn)
	defer close(doneCh)

	select {
	case <-calledCh:
		t.Fatal("expected callback to not invoke")
	case <-time.After(time.Second):
	}
}

func TestRandomStagger(t *testing.T) {
	intv := 24 * time.Hour
	min := 18 * time.Hour
	max := 30 * time.Hour
	for i := 0; i < 1000; i++ {
		out := randomStagger(intv)
		if out < min || out > max {
			t.Fatalf("unexpected value: %v", out)
		}
	}
}
