package config

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/SomniSom/docker-ops/internal/locale"
)

var (
	reYAMLLine    = regexp.MustCompile(`(?i)line\s+(\d+)`)
	reYAMLColumn  = regexp.MustCompile(`(?i)column\s+(\d+)`)
	reIndentedKey = regexp.MustCompile(`^([ \t]+)([a-zA-Z_][a-zA-Z0-9_]*)\s*:`)
)

// FormatYAMLParseError turns a low-level yaml parse error into a readable message with context.
func FormatYAMLParseError(filePath string, content []byte, parseErr error) error {
	if parseErr == nil {
		return nil
	}
	rel := filepath.Base(filePath)
	lines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
	msg := parseErr.Error()

	lineNum := extractLineNumber(msg)
	colNum := extractColumnNumber(msg)

	var b strings.Builder
	b.WriteString(locale.Tf("yaml.invalid_intro", rel))
	b.WriteString(locale.Tf("yaml.parser_said", msg))

	if extra := scanMisindentedRootKeys(lines); extra != "" {
		fmt.Fprintf(&b, "\n%s\n", extra)
	}

	if lineNum >= 1 && lineNum <= len(lines) {
		b.WriteString(locale.Tf("yaml.context_header", lineNum))
		printContext(&b, lines, lineNum, colNum)
		hint := hintForBadLine(lines, lineNum)
		if hint != "" {
			fmt.Fprintf(&b, "\n%s\n", hint)
		}
		// go-yaml often reports a wrong line; show hint for other suspicious lines too
		for i := 1; i <= len(lines); i++ {
			if i == lineNum {
				continue
			}
			if h := hintForBadLine(lines, i); h != "" && strings.Contains(h, locale.T("yaml.hint.top_level")) {
				fmt.Fprintf(&b, "\n%s\n", h)
				printContext(&b, lines, i, 0)
				break
			}
		}
	} else if lineNum > len(lines) && len(lines) > 0 {
		b.WriteString(locale.Tf("yaml.past_eof", lineNum))
		start := len(lines) - 4
		if start < 1 {
			start = 1
		}
		printContext(&b, lines, len(lines), 0)
		_ = start
	} else {
		b.WriteString(locale.T("yaml.no_line"))
		b.WriteString(locale.T("yaml.common_issues"))
		b.WriteString(locale.T("yaml.issue.root_keys"))
		b.WriteString(locale.T("yaml.issue.spaces_lists"))
		b.WriteString(locale.T("yaml.issue.colon_quote"))
	}

	return fmt.Errorf("%s", b.String())
}

// knownRootYAMLKeys must be unindented at column 0 in docker-ops.yml.
var knownRootYAMLKeys = []string{
	"compose_project_name", "compose_file", "compose_service",
	"remote_ssh", "remote_path", "ssh_identity",
	"exclude", "rsync_extra",
	"deploy_mode", "deploy_image", "deploy_push",
	"deploy_use_registry", "deploy_save_load", "deploy_save_compress", "deploy_build_remote",
	"deploy_include", "app_config", "help_show_effective", "use_remote",
}

func scanMisindentedRootKeys(lines []string) string {
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		m := reIndentedKey.FindStringSubmatch(line)
		if len(m) < 3 {
			continue
		}
		key := m[2]
		if !isKnownRootKey(key) {
			continue
		}
		return locale.Tf("yaml.misindented_root", key, i+1)
	}
	return ""
}

func isKnownRootKey(key string) bool {
	for _, k := range knownRootYAMLKeys {
		if k == key {
			return true
		}
	}
	return false
}

func extractLineNumber(msg string) int {
	m := reYAMLLine.FindStringSubmatch(msg)
	if len(m) < 2 {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

func extractColumnNumber(msg string) int {
	m := reYAMLColumn.FindStringSubmatch(msg)
	if len(m) < 2 {
		return 0
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

func printContext(b *strings.Builder, lines []string, lineNum, colNum int) {
	const margin = 2
	start := lineNum - margin
	if start < 1 {
		start = 1
	}
	end := lineNum + margin
	if end > len(lines) {
		end = len(lines)
	}
	width := len(strconv.Itoa(end))
	for i := start; i <= end; i++ {
		prefix := " "
		if i == lineNum {
			prefix = ">"
		}
		fmt.Fprintf(b, "%s %*d | %s\n", prefix, width, i, lines[i-1])
	}
	if colNum > 0 && lineNum >= 1 && lineNum <= len(lines) {
		line := lines[lineNum-1]
		// caret under the column (1-based)
		pad := width + 4 + colNum - 1
		if pad > 0 && colNum <= len(line)+1 {
			fmt.Fprintf(b, "%s^\n", strings.Repeat(" ", pad))
		}
	}
}

func hintForBadLine(lines []string, lineNum int) string {
	if lineNum < 1 || lineNum > len(lines) {
		return ""
	}
	line := lines[lineNum-1]
	m := reIndentedKey.FindStringSubmatch(line)
	if m != nil {
		indent := m[1]
		key := m[2]
		if strings.Contains(indent, "\t") {
			return locale.Tf("yaml.hint.tabs", lineNum, key)
		}
		if len(indent) > 0 {
			// Heuristic: root keys in our schema are snake_case at column 0
			return locale.Tf("yaml.hint.spaces", key, len(indent), key)
		}
	}
	trim := strings.TrimSpace(line)
	if trim != "" && !strings.HasPrefix(trim, "#") && strings.Contains(line, ":") {
		before, _, ok := strings.Cut(trim, ":")
		if ok && strings.Contains(before, " ") && !strings.HasPrefix(before, "\"") {
			return locale.Tf("yaml.hint.key_space", before)
		}
	}
	return ""
}
