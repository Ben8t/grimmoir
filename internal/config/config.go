package config

import (
	"os"
	"path/filepath"
)

func StorePath() string {
	if p := os.Getenv("GRIMMOIR_PATH"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".grimmoir"
	}
	return filepath.Join(home, ".grimmoir")
}
