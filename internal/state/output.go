package state

import (
	"os"
	"path/filepath"
)

// CreateOutput refuses overwrites and protects the file with native permissions.
func CreateOutput(path string) (*os.File, error) {
	return createOutput(path)
}

// CreateTempOutput creates a private, unpredictable scratch file from birth.
func CreateTempOutput(directory, prefix string) (*os.File, error) {
	random, err := RandomKey()
	if err != nil {
		return nil, err
	}
	return CreateOutput(filepath.Join(directory, prefix+Digest(random)+".tmp"))
}
