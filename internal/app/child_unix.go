//go:build linux || darwin

package app

import "os/exec"

func configureChild(*exec.Cmd) {}
