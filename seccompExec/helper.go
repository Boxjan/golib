package seccompExec

// +build !windows

import (
	"github.com/Boxjan/golib/seccompExec/scmpSyscall"
	"syscall"
)

func environForSysProcAttr(sys *scmpSyscall.SysProcAttr) ([]string, error) {
	return syscall.Environ(), nil
}
