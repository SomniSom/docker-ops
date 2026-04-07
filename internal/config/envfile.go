package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/SomniSom/docker-ops/internal/locale"
)

// ParseDQEnv reads dq.env (dotenv-style). Empty values are omitted so they do not override YAML (§14.1).
func ParseDQEnv(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	out := make(map[string]string)
	sc := bufio.NewScanner(f)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("%s", locale.Tf("envfile.expected_kv", path, lineNo))
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		// Strip optional surrounding quotes
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if key == "" {
			return nil, fmt.Errorf("%s", locale.Tf("envfile.empty_key", path, lineNo))
		}
		if val == "" {
			continue
		}
		out[key] = val
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
