//go:build windows

package main

import (
	"os"
	"syscall"
)

var (
	kernel32          = syscall.NewLazyDLL("kernel32.dll")
	procAttachConsole = kernel32.NewProc("AttachConsole")
)

const (
	ATTACH_PARENT_PROCESS = ^uint32(0) // 0xFFFFFFFF
)

func isHandleRedirected(stdHandle int) bool {
	h, err := syscall.GetStdHandle(stdHandle)
	if err != nil || h == syscall.InvalidHandle || h == 0 {
		return false
	}
	fileType, err := syscall.GetFileType(h)
	if err != nil {
		return false
	}
	return fileType == syscall.FILE_TYPE_PIPE || fileType == syscall.FILE_TYPE_DISK
}

func initConsole() {
	// Check if stdin, stdout, or stderr are already redirected (e.g. MCP client spawned us with pipes)
	stdinRedirected := isHandleRedirected(syscall.STD_INPUT_HANDLE)
	stdoutRedirected := isHandleRedirected(syscall.STD_OUTPUT_HANDLE)
	stderrRedirected := isHandleRedirected(syscall.STD_ERROR_HANDLE)

	// If both stdin and stdout are already redirected (pipes or files), keep them intact.
	// We MUST NOT attach to the parent console or reopen CONIN$/CONOUT$, as doing so
	// would hijack the MCP communication pipes and drop JSON-RPC traffic.
	if stdinRedirected && stdoutRedirected {
		return
	}

	// If dbridge.exe is run as a GUI subsystem process interactively from CMD/PowerShell without redirection,
	// attach to the parent console so interactive typing and output works.
	r, _, _ := procAttachConsole.Call(uintptr(ATTACH_PARENT_PROCESS))
	if r != 0 {
		if !stdinRedirected {
			if stdin, err := os.OpenFile("CONIN$", os.O_RDONLY, 0); err == nil {
				os.Stdin = stdin
			}
		}
		if !stdoutRedirected {
			if stdout, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0); err == nil {
				os.Stdout = stdout
			}
		}
		if !stderrRedirected {
			if stderr, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0); err == nil {
				os.Stderr = stderr
			}
		}
	}
}
