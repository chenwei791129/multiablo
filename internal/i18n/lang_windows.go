//go:build windows

package i18n

import (
	"syscall"
)

var (
	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
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
		if subLangID == 0x01 { // Traditional Chinese (Taiwan, Hong Kong, Macau)
			return LangZhTW
		}
		// For Simplified Chinese (0x02) or other Chinese variants,
		// fall back to English since we don't have zh_CN translations yet.
		return LangEnUS
	default:
		return LangEnUS
	}
}
