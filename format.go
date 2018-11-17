package checkpoint

import (
	"context"
	"os"
	"path/filepath"
	"time"
)

// HomeDir returns the current users home directory irrespecitve of the OS
func HomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// ConfigDir returns the config directory for solo.io
func ConfigDir() (string, error) {
	d := filepath.Join(HomeDir(), ".soloio")
	_, err := os.Stat(d)
	if err == nil {
		return d, nil
	}
	if os.IsNotExist(err) {
		err = os.Mkdir(d, 0755)
		if err != nil {
			return "", err
		}
		return d, nil
	}

	return d, err
}

// Format1 calls a basic version check
func Format1(product string, version string, t time.Time) {
	sigfile := filepath.Join(HomeDir(), ".soloio.sig")
	configDir, err := ConfigDir()
	if err == nil {
		sigfile = filepath.Join(configDir, "soloio.sig")
	}
	ctx := context.Background()
	report := &ReportParams{
		Product:       product,
		Version:       version,
		StartTime:     t,
		EndTime:       time.Now(),
		SignatureFile: sigfile,
	}
	Report(ctx, report)
}
