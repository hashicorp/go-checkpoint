package checkpoint

import (
	"reflect"
	"testing"
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
	}

	actual, err := Check(&CheckParams{
		Product: "test",
		Version: "1.0",
	})

	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}
}
