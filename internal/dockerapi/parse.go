package dockerapi

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	reExecStartUnix = regexp.MustCompile(`-H\s+(?:unix://)?(/[^\s]+)`)
	reListenStream  = regexp.MustCompile(`ListenStream=(.+)`)
)

// ParseDockerHost extracts a unix filesystem path from DOCKER_HOST or docker context host.
func ParseDockerHost(host string) (path string, ok bool) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", false
	}
	if strings.HasPrefix(host, "unix://") {
		u, err := url.Parse(host)
		if err != nil || u.Path == "" {
			return strings.TrimPrefix(host, "unix://"), true
		}
		return u.Path, true
	}
	if strings.HasPrefix(host, "/") {
		return host, true
	}
	return "", false
}

// ExpandListenStream expands systemd ListenStream specifiers (%t, %i).
func ExpandListenStream(listen, runtimeDir string, uid int) string {
	listen = strings.TrimSpace(listen)
	if listen == "" {
		return ""
	}
	if strings.HasPrefix(listen, "/") {
		return listen
	}
	rt := strings.TrimSpace(runtimeDir)
	if rt == "" && uid >= 0 {
		rt = "/run/user/" + itoa(uid)
	}
	listen = strings.ReplaceAll(listen, "%t", rt)
	if uid >= 0 {
		listen = strings.ReplaceAll(listen, "%i", itoa(uid))
	}
	if strings.HasPrefix(listen, "/") {
		return listen
	}
	if rt != "" {
		return rt + "/" + strings.TrimPrefix(listen, "/")
	}
	return listen
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// ParseSystemdShow extracts ListenStream path from systemctl show output.
func ParseSystemdShow(out, runtimeDir string, uid int) (path string, ok bool) {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if m := reListenStream.FindStringSubmatch(line); len(m) == 2 {
			p := ExpandListenStream(m[1], runtimeDir, uid)
			if p != "" {
				return p, true
			}
		}
		if strings.HasPrefix(line, "ExecStart=") {
			if m := reExecStartUnix.FindStringSubmatch(line); len(m) == 2 {
				return m[1], true
			}
		}
	}
	return "", false
}
