package deploy

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/SomniSom/docker-ops/internal/config"
)

// DefaultExcludePatterns mirror the historical rsync/SFTP exclude list (readme §4.3).
var DefaultExcludePatterns = []string{
	".git/",
	"data/",
	"*.db",
	"*.db-shm",
	"*.db-wal",
	".cursor/",
	".env",
	"dq.env",
}

// MergeExcludePatterns returns defaults plus cfg.Exclude (YAML / merged config).
func MergeExcludePatterns(cfg *config.Config) []string {
	out := append([]string(nil), DefaultExcludePatterns...)
	if cfg != nil {
		for _, e := range cfg.Exclude {
			e = strings.TrimSpace(e)
			if e != "" {
				out = append(out, e)
			}
		}
	}
	return out
}

// PathExcluded reports whether rel (slash-separated, relative to project root, no leading "./") matches any pattern.
func PathExcluded(rel string, patterns []string) bool {
	rel = filepath.ToSlash(rel)
	rel = strings.TrimPrefix(rel, "./")
	if rel == "" || rel == "." {
		return false
	}
	base := path.Base(rel)
	segs := strings.Split(rel, "/")
	for _, pat := range patterns {
		pat = filepath.ToSlash(strings.TrimSpace(pat))
		if pat == "" {
			continue
		}
		if strings.HasSuffix(pat, "/") {
			dir := strings.TrimSuffix(pat, "/")
			if dir == "" {
				continue
			}
			for _, s := range segs {
				if s == dir {
					return true
				}
			}
			if rel == dir || strings.HasPrefix(rel, dir+"/") {
				return true
			}
			continue
		}
		if strings.Contains(pat, "*") || strings.Contains(pat, "?") || strings.Contains(pat, "[") {
			if ok, _ := path.Match(pat, base); ok {
				return true
			}
			if ok, _ := path.Match(pat, rel); ok {
				return true
			}
			continue
		}
		if base == pat || rel == pat {
			return true
		}
	}
	return false
}
