//go:build linux

package db

import (
	"os"
)

func DBPath() (string, error) {
	dir, err := os.UserHomeDir()
	// TODO: find path for master.mdb on Linux Steam installs
	return dir + "/", err
}
