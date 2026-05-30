package dockerapi

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/locale"
	"github.com/SomniSom/docker-ops/internal/sshexec"
	"golang.org/x/crypto/ssh"
)

// ResolvedSocket holds auto-detected remote docker socket path.
type ResolvedSocket struct {
	Path   string
	Source string
}

// ResolveRemoteSocket determines unix socket path on remote host via SSH.
func ResolveRemoteSocket(sshClient *ssh.Client, cfg *config.Config) (ResolvedSocket, error) {
	if sshClient == nil {
		return ResolvedSocket{}, fmt.Errorf("ssh client is nil")
	}
	if p, ok := config.RemoteDockerSocketExplicit(cfg); ok {
		return ResolvedSocket{Path: p, Source: "config"}, nil
	}
	if r, ok, err := resolveFromRemoteEnv(sshClient); err == nil && ok {
		return r, nil
	} else if err != nil {
		return ResolvedSocket{}, err
	}
	if r, ok, err := resolveFromDockerContext(sshClient); err == nil && ok {
		return r, nil
	} else if err != nil {
		return ResolvedSocket{}, err
	}
	if r, ok, err := resolveFromSystemd(sshClient, cfg); err == nil && ok {
		return r, nil
	} else if err != nil {
		return ResolvedSocket{}, err
	}
	return resolveByProbe(sshClient)
}

func runRemoteOut(sshClient *ssh.Client, script string) (string, error) {
	sess, err := sshClient.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()
	out, err := sess.CombinedOutput("bash -lc " + sshexec.ShellQuote(script))
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func resolveFromRemoteEnv(sshClient *ssh.Client) (ResolvedSocket, bool, error) {
	out, err := runRemoteOut(sshClient, `echo "${DOCKER_HOST:-}"`)
	if err != nil {
		return ResolvedSocket{}, false, nil
	}
	if p, ok := ParseDockerHost(strings.TrimSpace(out)); ok {
		return ResolvedSocket{Path: p, Source: "DOCKER_HOST"}, true, nil
	}
	return ResolvedSocket{}, false, nil
}

func resolveFromDockerContext(sshClient *ssh.Client) (ResolvedSocket, bool, error) {
	out, err := runRemoteOut(sshClient, `command -v docker >/dev/null 2>&1 && docker context show --format '{{.Endpoints.docker.Host}}' 2>/dev/null || true`)
	if err != nil {
		return ResolvedSocket{}, false, nil
	}
	if p, ok := ParseDockerHost(strings.TrimSpace(out)); ok {
		return ResolvedSocket{Path: p, Source: "docker context"}, true, nil
	}
	return ResolvedSocket{}, false, nil
}

func resolveFromSystemd(sshClient *ssh.Client, cfg *config.Config) (ResolvedSocket, bool, error) {
	units := []string{}
	if cfg != nil && strings.TrimSpace(cfg.RemoteDockerSocketService) != "" {
		units = append(units, strings.TrimSpace(cfg.RemoteDockerSocketService))
	} else {
		units = append(units, "docker.socket", "docker.service")
	}
	meta, _ := runRemoteOut(sshClient, `echo "UID=$(id -u)"; echo "XDG_RUNTIME_DIR=${XDG_RUNTIME_DIR:-}"`)
	uid := 0
	rt := ""
	for _, line := range strings.Split(meta, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "UID=") {
			fmt.Sscanf(strings.TrimPrefix(line, "UID="), "%d", &uid)
		}
		if strings.HasPrefix(line, "XDG_RUNTIME_DIR=") {
			rt = strings.TrimPrefix(line, "XDG_RUNTIME_DIR=")
		}
	}
	for _, unit := range units {
		script := fmt.Sprintf(`systemctl show -p ListenStream,ExecStart %q 2>/dev/null || systemctl --user show -p ListenStream,ExecStart %q 2>/dev/null || true`, unit, unit)
		out, err := runRemoteOut(sshClient, script)
		if err != nil || strings.TrimSpace(out) == "" {
			continue
		}
		if p, ok := ParseSystemdShow(out, rt, uid); ok {
			return ResolvedSocket{Path: p, Source: "systemd " + unit}, true, nil
		}
	}
	return ResolvedSocket{}, false, nil
}

func resolveByProbe(sshClient *ssh.Client) (ResolvedSocket, error) {
	script := `candidates=("/var/run/docker.sock" "/run/docker.sock" "/run/user/$(id -u)/docker.sock")
if [ -n "${XDG_RUNTIME_DIR:-}" ]; then candidates+=("${XDG_RUNTIME_DIR}/docker.sock"); fi
found=""
for p in "${candidates[@]}"; do
  if [ -S "$p" ]; then
    echo "$p"
    found=1
  fi
done
if [ -z "$found" ]; then exit 42; fi`
	out, err := runRemoteOut(sshClient, script)
	if err != nil {
		if strings.Contains(err.Error(), "exit status 42") {
			return ResolvedSocket{}, fmt.Errorf("%s", locale.T("dockerapi.err.no_socket"))
		}
		return ResolvedSocket{}, err
	}
	var paths []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			paths = append(paths, line)
		}
	}
	if len(paths) == 0 {
		return ResolvedSocket{}, fmt.Errorf("%s", locale.T("dockerapi.err.no_socket"))
	}
	chosen := paths[0]
	source := "probe"
	if len(paths) > 1 {
		fmt.Fprint(os.Stderr, locale.Tf("dockerapi.warn.multiple_sockets", strings.Join(paths, ", ")))
		for _, p := range paths {
			if strings.Contains(p, "/run/user/") {
				chosen = p
				source = "probe (user socket preferred)"
				break
			}
		}
	}
	return ResolvedSocket{Path: chosen, Source: source}, nil
}

// RemoteSession holds SSH tunnel and docker API client for remote host.
type RemoteSession struct {
	Tunnel *Tunnel
	Client *Client
	Socket ResolvedSocket
}

// OpenRemote opens tunnel + docker client after resolving socket.
func OpenRemote(ctx context.Context, sshClient *ssh.Client, cfg *config.Config) (*RemoteSession, error) {
	sock, err := ResolveRemoteSocket(sshClient, cfg)
	if err != nil {
		return nil, err
	}
	fmt.Fprint(os.Stderr, locale.Tf("dockerapi.socket_resolved", sock.Path, sock.Source))
	tun, err := StartTunnel(sshClient, sock.Path)
	if err != nil {
		return nil, err
	}
	cli, err := NewFromHost(ctx, tun.Host())
	if err != nil {
		_ = tun.Close()
		return nil, err
	}
	return &RemoteSession{Tunnel: tun, Client: cli, Socket: sock}, nil
}

// Close closes client and tunnel.
func (s *RemoteSession) Close() error {
	if s == nil {
		return nil
	}
	var err error
	if s.Client != nil {
		if e := s.Client.Close(); e != nil {
			err = e
		}
	}
	if s.Tunnel != nil {
		if e := s.Tunnel.Close(); e != nil && err == nil {
			err = e
		}
	}
	return err
}

// TryOpenRemote opens remote API session; returns nil session on failure (for optional optimization).
func TryOpenRemote(ctx context.Context, sshClient *ssh.Client, cfg *config.Config) (*RemoteSession, error) {
	if sshClient == nil {
		return nil, fmt.Errorf("ssh client is nil")
	}
	return OpenRemote(ctx, sshClient, cfg)
}
