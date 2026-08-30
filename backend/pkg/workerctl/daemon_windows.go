//go:build windows

package workerctl

import "syscall"

func isProcessRunning(pid int) bool {
	h, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(h)

	var code uint32
	if err := syscall.GetExitCodeProcess(h, &code); err != nil {
		return false
	}
	const stillActive = 259
	return code == stillActive
}

func sysProcAttrDetached() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    true,
	}
}
