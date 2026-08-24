//go:build windows

package tui

import "golang.org/x/sys/windows"

const virtualKeyControl = 0x11

var getAsyncKeyState = windows.NewLazySystemDLL("user32.dll").NewProc("GetAsyncKeyState")

func platformControlKeyPressed() bool {
	state, _, _ := getAsyncKeyState.Call(virtualKeyControl)
	return state&0x8000 != 0
}
