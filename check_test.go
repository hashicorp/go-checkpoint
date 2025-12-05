// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package checkpoint

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// roundTripFunc lets us create a custom function to handle HTTP requests, making it easy 
// to mock network responses in tests by implementing the http.RoundTripper interface.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (rtf roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rtf(req)
}

func TestCheck(t *testing.T) {
	expected := &CheckResponse{
		Product:             "test",
		CurrentVersion:      "1.0.2",
		CurrentReleaseDate:  0,
		CurrentDownloadURL:  "http://www.hashicorp.com/",
		CurrentChangelogURL: "http://www.hashicorp.com/",
		ProjectWebsite:      "http://www.hashicorp.com",
		Outdated:            true,
		Alerts:              []*CheckAlert{},
	}

	// Mock HTTP client to return the expected response
	mockResp := `{
		"product": "test",
		"current_version": "1.0.2",
		"current_release_date": 0,
		"current_download_url": "http://www.hashicorp.com/",
		"current_changelog_url": "http://www.hashicorp.com/",
		"project_website": "http://www.hashicorp.com",
		"outdated": true,
		"alerts": []
	}`
	mockClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(mockResp)),
				Header:     make(http.Header),
			}, nil
		}),
	}
	actual, err := Check(&CheckParams{
		Product:    "test",
		Version:    "1.0",
		HTTPClient: mockClient,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected %#v, got: %#v", expected, actual)
	}
}

func TestCheck_timeout(t *testing.T) {
	if err := os.Setenv("CHECKPOINT_TIMEOUT", "50"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	defer func() {
		if err := os.Setenv("CHECKPOINT_TIMEOUT", ""); err != nil {
			t.Fatalf("failed to reset env: %v", err)
		}
	}()

	expected := "Client.Timeout exceeded while awaiting headers"

	_, err := Check(&CheckParams{
		Product: "test",
		Version: "1.0",
	})

	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected a timeout error, got: %v", err)
	}
}

func TestCheck_disabled(t *testing.T) {
	if err := os.Setenv("CHECKPOINT_DISABLE", "1"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	defer func() {
		if err := os.Setenv("CHECKPOINT_DISABLE", ""); err != nil {
			t.Fatalf("failed to reset env: %v", err)
		}
	}()

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
	dir, err := os.MkdirTemp("", "checkpoint")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := &CheckResponse{
		Product:             "test",
		CurrentVersion:      "1.0.2",
		CurrentReleaseDate:  0,
		CurrentDownloadURL:  "http://www.hashicorp.com/",
		CurrentChangelogURL: "http://www.hashicorp.com/",
		ProjectWebsite:      "http://www.hashicorp.com",
		Outdated:            true,
		Alerts:              []*CheckAlert{},
	}

	// Mock HTTP client to return the expected response
	mockResp := `{
		"product": "test",
		"current_version": "1.0.2",
		"current_release_date": 0,
		"current_download_url": "http://www.hashicorp.com/",
		"current_changelog_url": "http://www.hashicorp.com/",
		"project_website": "http://www.hashicorp.com",
		"outdated": true,
		"alerts": []
	}`
	mockClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(mockResp)),
				Header:     make(http.Header),
			}, nil
		}),
	}
	var actual *CheckResponse
	for i := 0; i < 5; i++ {
		var err error
		actual, err = Check(&CheckParams{
			Product:    "test",
			Version:    "1.0",
			CacheFile:  filepath.Join(dir, "cache"),
			HTTPClient: mockClient,
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
	dir, err := os.MkdirTemp("", "checkpoint")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expected := &CheckResponse{
		Product:             "test",
		CurrentVersion:      "1.0.2",
		CurrentReleaseDate:  0,
		CurrentDownloadURL:  "http://www.hashicorp.com/",
		CurrentChangelogURL: "http://www.hashicorp.com/",
		ProjectWebsite:      "http://www.hashicorp.com",
		Outdated:            true,
		Alerts:              []*CheckAlert{},
	}

	// Mock HTTP client to return the expected response
	mockResp := `{
		"product": "test",
		"current_version": "1.0.2",
		"current_release_date": 0,
		"current_download_url": "http://www.hashicorp.com/",
		"current_changelog_url": "http://www.hashicorp.com/",
		"project_website": "http://www.hashicorp.com",
		"outdated": true,
		"alerts": []
	}`
	mockClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(mockResp)),
				Header:     make(http.Header),
			}, nil
		}),
	}
	var actual *CheckResponse
	for i := 0; i < 5; i++ {
		var err error
		actual, err = Check(&CheckParams{
			Product:    "test",
			Version:    "1.0",
			CacheFile:  filepath.Join(dir, "nested", "cache"),
			HTTPClient: mockClient,
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
		CurrentVersion:      "1.0.2",
		CurrentReleaseDate:  0,
		CurrentDownloadURL:  "http://www.hashicorp.com/",
		CurrentChangelogURL: "http://www.hashicorp.com/",
		ProjectWebsite:      "http://www.hashicorp.com",
		Outdated:            true,
		Alerts:              []*CheckAlert{},
	}

	params := &CheckParams{
		Product: "test",
		Version: "1.0",
	}

	// Mock HTTP client to return the expected response
	mockResp := `{
		"product": "test",
		"current_version": "1.0.2",
		"current_release_date": 0,
		"current_download_url": "http://www.hashicorp.com/",
		"current_changelog_url": "http://www.hashicorp.com/",
		"project_website": "http://www.hashicorp.com",
		"outdated": true,
		"alerts": []
	}`
	mockClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(mockResp)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	calledCh := make(chan struct{})
	checkFn := func(actual *CheckResponse, err error) {
		// We check if calledCh is already closed before closing it to avoid a panic from 
		// double-closing the channel.
		defer func() {
			select {
			case <-calledCh:
				// already closed
			default:
				close(calledCh)
			}
		}()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("expected %#v, got: %#v", expected, actual)
		}
	}

	params.HTTPClient = mockClient
	doneCh := CheckInterval(params, 10*time.Millisecond, checkFn)
	defer close(doneCh)

	select {
	case <-calledCh:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout")
	}
}

func TestCheckInterval_disabled(t *testing.T) {
	if err := os.Setenv("CHECKPOINT_DISABLE", "1"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	defer func() {
		if err := os.Setenv("CHECKPOINT_DISABLE", ""); err != nil {
			t.Fatalf("failed to reset env: %v", err)
		}
	}()

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
