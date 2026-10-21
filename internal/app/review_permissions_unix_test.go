//go:build linux || darwin

package app

import "os"

func exposeFixture(path string) error { return os.Chmod(path, 0644) }
