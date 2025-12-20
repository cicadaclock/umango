//go:build windows

package db

import (
	"os"
)

func DBPath() (string, error) {
	dir, err := os.UserHomeDir()
	return dir + "\\AppData\\LocalLow\\Cygames\\Umamusume\\master\\master.mdb", err
}
