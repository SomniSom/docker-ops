package locale

import (
	"os"
	"strings"
)

// BootstrapFromArgs sets language from DQ_LANG, then --lang anywhere in argv, else auto-detect.
func BootstrapFromArgs(argv []string) {
	if v := strings.TrimSpace(os.Getenv("DQ_LANG")); v != "" {
		Set(v)
		return
	}
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		if strings.HasPrefix(a, "--lang=") {
			Set(strings.TrimPrefix(a, "--lang="))
			return
		}
		if a == "--lang" && i+1 < len(argv) {
			next := argv[i+1]
			if next != "" && !strings.HasPrefix(next, "-") {
				Set(next)
				return
			}
		}
	}
	Set("auto")
}
