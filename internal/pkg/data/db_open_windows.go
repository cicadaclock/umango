//go:build windows

package data

import (
	"os"
)

func DBPath() (string, error) {
	dir, err := os.UserHomeDir()
	return dir + "\\AppData\\LocalLow\\Cygames\\Umamusume\\master\\master.mdb", err
}
