// Package marker manages the run-once sentinel file that prevents the
// onboarding wizard from showing again after completion.
package marker

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	appDir   = "day1"
	fileName = ".completed"
)

func path() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	return filepath.Join(base, appDir, fileName), nil
}

func Exists() (bool, error) {
	p, err := path()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(p)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func Write() error {
	p, err := path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return os.WriteFile(p, []byte(time.Now().UTC().Format(time.RFC3339)+"\n"), 0o644)
}

func Remove() error {
	p, err := path()
	if err != nil {
		return err
	}
	err = os.Remove(p)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
