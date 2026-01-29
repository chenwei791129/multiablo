//go:build !windows

package i18n

import (
	"os"
	"strings"
)

// detectSystemLanguage detects the system language on non-Windows systems.
func detectSystemLanguage() string {
	// Check LANG environment variable
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LC_ALL")
	}
	if lang == "" {
		lang = os.Getenv("LC_MESSAGES")
	}

	if lang == "" {
		return LangEnUS
	}

	// Parse language code (e.g., "zh_TW.UTF-8" -> "zh_TW")
	lang = strings.Split(lang, ".")[0]

	switch {
	case strings.HasPrefix(lang, "zh_TW"):
		return LangZhTW
	case strings.HasPrefix(lang, "zh_Hant"):
		return LangZhTW
	default:
		return LangEnUS
	}
}
