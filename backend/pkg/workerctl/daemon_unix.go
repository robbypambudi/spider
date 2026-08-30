//go:build !windows

package workerctl

import "syscall"

func isProcessRunning(pid int) bool {
	// FindProcess never fails on Unix; signal 0 does the actual existence check.
	return syscall.Kill(pid, 0) == nil
}

func sysProcAttrDetached() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
