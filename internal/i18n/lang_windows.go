//go:build windows

package i18n

import (
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32              = syscall.NewLazyDLL("kernel32.dll")
	procGetUserDefaultUILanguage = kernel32.NewProc("GetUserDefaultUILanguage")
)

// detectSystemLanguage detects the system UI language on Windows.
func detectSystemLanguage() string {
	langID, _, _ := procGetUserDefaultUILanguage.Call()

	// Extract primary language ID (lower 10 bits)
	primaryLangID := langID & 0x3FF

	switch primaryLangID {
	case 0x04: // Chinese
		// Check sublanguage for Traditional vs Simplified
		subLangID := (langID >> 10) & 0x3F
		if subLangID == 0x01 { // Traditional Chinese
			return LangZhTW
		}
		// Default to Traditional Chinese for any Chinese variant
		return LangZhTW
	default:
		return LangEnUS
	}
}

// getLocaleInfo retrieves locale information from Windows.
func getLocaleInfo(lcid uint32, lcType uint32) string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getLocaleInfoW := kernel32.NewProc("GetLocaleInfoW")

	// First call to get required buffer size
	size, _, _ := getLocaleInfoW.Call(
		uintptr(lcid),
		uintptr(lcType),
		0,
		0,
	)

	if size == 0 {
		return ""
	}

	// Allocate buffer and get the actual value
	buf := make([]uint16, size)
	ret, _, _ := getLocaleInfoW.Call(
		uintptr(lcid),
		uintptr(lcType),
		uintptr(unsafe.Pointer(&buf[0])),
		size,
	)

	if ret == 0 {
		return ""
	}

	return strings.TrimRight(syscall.UTF16ToString(buf), "\x00")
}
