package fsutil

import "os"

// Exists returns a boolean indicating whether or not the path is present. It
// panics on errors.
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}

		panic(err)
	}

	return true
}
