//go:build windows

package console

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	systemMetricScreenWidth  = 0
	systemMetricScreenHeight = 1
)

type rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

var (
	kernel32             = windows.NewLazySystemDLL("kernel32.dll")
	user32               = windows.NewLazySystemDLL("user32.dll")
	procGetConsoleWindow = kernel32.NewProc("GetConsoleWindow")
	procGetWindowRect    = user32.NewProc("GetWindowRect")
	procGetSystemMetrics = user32.NewProc("GetSystemMetrics")
	procMoveWindow       = user32.NewProc("MoveWindow")
)

func CenterWindow() {
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		return
	}

	var bounds rect
	ok, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&bounds)))
	if ok == 0 {
		return
	}

	width := bounds.Right - bounds.Left
	height := bounds.Bottom - bounds.Top
	if width <= 0 || height <= 0 {
		return
	}

	screenWidth, _, _ := procGetSystemMetrics.Call(systemMetricScreenWidth)
	screenHeight, _, _ := procGetSystemMetrics.Call(systemMetricScreenHeight)
	if screenWidth == 0 || screenHeight == 0 {
		return
	}

	x := (int32(screenWidth) - width) / 2
	y := (int32(screenHeight) - height) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	procMoveWindow.Call(
		hwnd,
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		uintptr(1),
	)
}
