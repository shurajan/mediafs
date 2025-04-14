package service

import (
	"os"
)

func IsTestMode() bool {
	if os.Getenv("MEDIAFS_MODE") == "test" {
		return true
	}
	for _, arg := range os.Args[1:] {
		if arg == "--test" {
			return true
		}
	}
	return false
}
