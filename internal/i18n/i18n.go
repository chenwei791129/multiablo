// Package i18n provides internationalization support for the application.
//
// Thread Safety: Init() and SetLanguage() must be called during application
// initialization before any goroutines that use Get() are started. Once
// initialized, the Get() function is safe to call from multiple goroutines
// as it only performs read operations on the locale data.
package i18n

import (
	"embed"
	"io/fs"
	"path"

	"github.com/leonelquinteros/gotext"
)

//go:embed locales
var localesFS embed.FS

const (
	// Domain is the translation domain name
	Domain = "default"
)

// Supported languages
const (
	LangEnUS = "en_US"
	LangZhTW = "zh_TW"
)

var (
	// currentLocale holds the active locale
	currentLocale *gotext.Locale
	// currentLang holds the current language code
	currentLang string
)

// Init initializes the i18n system with the specified language.
// If lang is empty, it will try to detect the system language.
//
// This function must be called once during application startup, before
// any goroutines that call Get() are started. It is not thread-safe.
func Init(lang string) {
	if lang == "" {
		lang = detectSystemLanguage()
	}

	currentLang = lang

	// Load embedded PO file
	poPath := path.Join("locales", lang, "LC_MESSAGES", Domain+".po")
	poData, err := localesFS.ReadFile(poPath)
	if err != nil {
		// Fallback to English if translation not found
		currentLang = LangEnUS
		poPath = path.Join("locales", LangEnUS, "LC_MESSAGES", Domain+".po")
		poData, err = localesFS.ReadFile(poPath)
		if err != nil {
			// No translation available, use original strings
			currentLocale = nil
			return
		}
	}

	// Create locale with embedded filesystem
	subFS, err := fs.Sub(localesFS, "locales")
	if err != nil {
		currentLocale = nil
		return
	}

	currentLocale = gotext.NewLocaleFS(currentLang, subFS)

	// Parse PO data and add as translator
	po := gotext.NewPo()
	po.Parse(poData)
	currentLocale.AddTranslator(Domain, po)
}

// Get returns the translated string for the given message ID.
// This function returns the translated string without any variable substitution.
// For format strings with variables, use fmt.Sprintf with Get as the format.
func Get(msgID string) string {
	if currentLocale == nil {
		return msgID
	}
	return getTranslation(msgID)
}

// getTranslation is a helper that performs the actual translation lookup.
// Separated to avoid go vet false positive about non-constant format string.
func getTranslation(msgID string) string {
	tr, ok := currentLocale.Domains[Domain]
	if !ok {
		return msgID
	}
	translated := tr.Get(msgID)
	if translated == "" {
		return msgID
	}
	return translated
}

// GetN returns the translated string with plural support.
func GetN(msgID, msgIDPlural string, n int, args ...interface{}) string {
	if currentLocale == nil {
		if n == 1 {
			return msgID
		}
		return msgIDPlural
	}
	return currentLocale.GetN(msgID, msgIDPlural, n, args...)
}

// GetCurrentLanguage returns the current language code.
func GetCurrentLanguage() string {
	return currentLang
}

// SetLanguage changes the current language.
func SetLanguage(lang string) {
	Init(lang)
}

// GetAvailableLanguages returns a list of supported language codes.
func GetAvailableLanguages() []string {
	return []string{LangEnUS, LangZhTW}
}
