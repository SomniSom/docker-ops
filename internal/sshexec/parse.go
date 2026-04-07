package sshexec

import (
	"fmt"
	"net"
	"strings"

	"github.com/SomniSom/docker-ops/internal/locale"
)

// ParseUserHost parses remote_ssh as user@host or user@host:port or user@[ipv6]:port.
func ParseUserHost(remoteSSH string) (user, address string, err error) {
	s := strings.TrimSpace(remoteSSH)
	if s == "" {
		return "", "", fmt.Errorf("%s", locale.T("ssh.err.empty_remote"))
	}
	at := strings.IndexByte(s, '@')
	if at <= 0 || at == len(s)-1 {
		return "", "", fmt.Errorf("%s", locale.Tf("ssh.err.bad_remote", remoteSSH))
	}
	user = s[:at]
	hostport := s[at+1:]
	if hostport == "" {
		return "", "", fmt.Errorf("%s", locale.T("ssh.err.no_host"))
	}
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		// no :port
		host = hostport
		port = "22"
	}
	address = net.JoinHostPort(host, port)
	return user, address, nil
}
