//go:build windows

package app

import (
	"os/exec"
	"syscall"
)

func configureChild(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
