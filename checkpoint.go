// checkpoint is a package for checking version information and alerts
// for a HashiCorp product.
package checkpoint

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
)

// CheckParams are the parameters for configuring a check request.
type CheckParams struct {
	// Product and version are used to lookup the correct product and
	// alerts for the proper version. The version is also used to perform
	// a version check.
	Product string
	Version string

	// Arch and OS are used to filter alerts potentially only to things
	// affecting a specific os/arch combination. If these aren't specified,
	// they'll be automatically filled in.
	Arch string
	OS   string

	// Signature is some random signature that should be stored and used
	// as a cookie-like value. This ensures that alerts aren't repeated.
	// If the signature is changed, repeat alerts may be sent down. The
	// signature should NOT be anything identifiable to a user (such as
	// a MAC address). It should be random.
	//
	// If SignatureFile is given, then the signature will be read from this
	// file. If the file doesn't exist, then a random signature will
	// automatically be generated and stored here. SignatureFile will be
	// ignored if Signature is given.
	Signature     string
	SignatureFile string
}

// CheckResponse is the response for a check request.
type CheckResponse struct {
	Product             string
	CurrentVersion      string `json:"current_version"`
	CurrentReleaseDate  int    `json:"current_release_date"`
	CurrentDownloadURL  string `json:"current_download_url"`
	CurrentChangelogURL string `json:"current_changelog_url"`
	ProjectWebsite      string `json:"project_website"`
	Outdated            bool   `json:"outdated"`
	Alerts              []*CheckAlert
}

// CheckAlert is a single alert message from a check request.
//
// These never have to be manually constructed, and are typically populated
// into a CheckResponse as a result of the Check request.
type CheckAlert struct {
	ID      int
	Date    int
	Message string
	URL     string
	Level   string
}

// Check checks for alerts and new version information.
func Check(p *CheckParams) (*CheckResponse, error) {
	var u url.URL

	if p.Arch == "" {
		p.Arch = runtime.GOARCH
	}
	if p.OS == "" {
		p.OS = runtime.GOOS
	}

	// If we're given a SignatureFile, then attempt to read that.
	signature := p.Signature
	if p.Signature == "" && p.SignatureFile != "" {
		var err error
		signature, err = checkSignature(p.SignatureFile)
		if err != nil {
			return nil, err
		}
	}

	v := u.Query()
	v.Set("version", p.Version)
	v.Set("arch", p.Arch)
	v.Set("os", p.OS)
	v.Set("signature", signature)

	u.Scheme = "http"
	u.Host = "api.checkpoint.hashicorp.com"
	u.Path = fmt.Sprintf("/v1/check/%s", p.Product)
	u.RawQuery = v.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Unknown status: %d", resp.StatusCode)
	}

	var result CheckResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func checkSignature(path string) (string, error) {
	_, err := os.Stat(path)
	if err == nil {
		// The file exists, read it out
		sigBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(string(sigBytes)), nil
	}

	// If this isn't a non-exist error, then return that.
	if !os.IsNotExist(err) {
		return "", err
	}

	// The file doesn't exist, so create a signature.
	var b [16]byte
	n := 0
	for n < 16 {
		n2, err := rand.Read(b[n:])
		if err != nil {
			return "", err
		}

		n += n2
	}
	signature := fmt.Sprintf(
		"%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	// Write the signature
	if err := ioutil.WriteFile(path, []byte(signature+"\n"), 0644); err != nil {
		return "", err
	}

	return signature, nil
}
