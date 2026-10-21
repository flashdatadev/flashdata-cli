//go:build linux || darwin

package app

func legacyProtected(b []byte) (string, []byte, error) { return "posix-private-file", b, nil }
