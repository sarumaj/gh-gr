package util

import "os"

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || (err != nil && os.IsNotExist(err))
}
