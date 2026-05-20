//go:build windows && !appengine

package terminal

import (
	"os"
	"syscall"
	"unsafe"
)

var (
    kernel32           = syscall.NewLazyDLL("kernel32.dll")
    procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
    procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

// enableVirtualTerminalProcessing enables ANSI escape code processing
// on Windows 10+. Without this, color codes print as literal text.
const enableVirtualTerminalProcessing = 0x0004

func init() {
    var mode uint32
    fd := os.Stdout.Fd()
    r, _, _ := syscall.SyscallN(procGetConsoleMode.Addr(), fd, uintptr(unsafe.Pointer(&mode)))
    if r == 0 {
        return
    }
    if (mode & enableVirtualTerminalProcessing) != enableVirtualTerminalProcessing {
        procSetConsoleMode.Call(fd, uintptr(mode|enableVirtualTerminalProcessing)) //nolint:errcheck
    }
}