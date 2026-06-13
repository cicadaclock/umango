//go:build windows

package db

import (
	"fmt"
	"os"
	"path/filepath"
)

func DBPath() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get user home dir: %w", err)
	}
	return filepath.Join(dir, "\\AppData\\LocalLow\\Cygames\\Umamusume\\master\\master.mdb"), nil
}
