// Package locale selects English (default) or Russian UI strings (readme §8).
package locale

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// Code is a supported UI language.
type Code int

const (
	En Code = iota
	Ru
)

var (
	mu   sync.RWMutex
	code = En
)

// Set switches UI language: "en", "ru", "auto" (from environment), or empty → auto.
func Set(s string) {
	mu.Lock()
	defer mu.Unlock()
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "ru", "rus", "russian":
		code = Ru
	case "en", "eng", "english":
		code = En
	case "auto", "":
		code = detectFromEnv()
	default:
		code = En
	}
}

// Current returns the active language code.
func Current() Code {
	mu.RLock()
	defer mu.RUnlock()
	return code
}

func detectFromEnv() Code {
	for _, k := range []string{"DQ_LANG", "LANGUAGE", "LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			if c := parseLocaleToken(v); c == Ru {
				return Ru
			}
		}
	}
	return En
}

// parseLocaleToken handles values like ru_RU.UTF-8, ru:en, C.UTF-8.
func parseLocaleToken(v string) Code {
	v = strings.ToLower(v)
	if i := strings.IndexAny(v, ".@"); i >= 0 {
		v = v[:i]
	}
	if i := strings.IndexByte(v, ':'); i >= 0 {
		v = v[:i]
	}
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, "ru") {
		return Ru
	}
	return En
}

// T returns the localized string for key, falling back to English, then to key.
func T(key string) string {
	mu.RLock()
	c := code
	mu.RUnlock()

	enStr, ok := catalogEn[key]
	if !ok {
		return key
	}
	if c == Ru {
		if ruStr, ok := catalogRu[key]; ok && ruStr != "" {
			return ruStr
		}
	}
	return enStr
}

// Tf formats T(key) with fmt.Sprintf.
func Tf(key string, args ...any) string {
	return fmt.Sprintf(T(key), args...)
}
