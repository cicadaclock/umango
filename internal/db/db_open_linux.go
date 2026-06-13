//go:build linux

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
	return filepath.Join(dir, "/.local/share/Steam/steamapps/compatdata/3224770/pfx/drive_c/users/steamuser/AppData/LocalLow/Cygames/Umamusume/master/master.mdb"), nil
}
