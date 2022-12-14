// This file was automatically generated by https://github.com/kbinani/win/blob/generator/internal/cmd/gen/gen.go
// go run internal/cmd/gen/gen.go

// +build windows

package ipkwiz

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

//ATOM atom
type ATOM uint16

var (
	// Library
	libkernel32 = windows.NewLazySystemDLL("kernel32.dll")

	// Functions
	globalAddAtom    = libkernel32.NewProc("GlobalAddAtomW")
	globalDeleteAtom = libkernel32.NewProc("GlobalDeleteAtom")
	globalFindAtom   = libkernel32.NewProc("GlobalFindAtomW")
)

//GlobalAddAtom WinAPI func
func GlobalAddAtom(lpString string) ATOM {
	lpStringStr, _ := windows.UTF16PtrFromString(lpString)

	r0, _, _ := globalAddAtom.Call(uintptr(unsafe.Pointer(lpStringStr)))
	return ATOM(r0)
}

//GlobalDeleteAtom WinAPI func
func GlobalDeleteAtom(nAtom ATOM) ATOM {
	r0, _, _ := globalDeleteAtom.Call(uintptr(nAtom))
	return ATOM(r0)
}

//GlobalFindAtom WinAPI func
func GlobalFindAtom(lpString string) ATOM {
	lpStringStr, _ := windows.UTF16PtrFromString(lpString)
	r0, _, _ := globalFindAtom.Call(uintptr(unsafe.Pointer(lpStringStr)))
	return ATOM(r0)
}
