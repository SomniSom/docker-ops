package dockerapi

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type streamLocalForwardMsg struct {
	SocketPath string
	Reserved   string
	Flags      uint32
}

type channelConn struct {
	ssh.Channel
}

func (c channelConn) LocalAddr() net.Addr              { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)} }
func (c channelConn) RemoteAddr() net.Addr             { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)} }
func (c channelConn) SetDeadline(t time.Time) error    { return nil }
func (c channelConn) SetReadDeadline(t time.Time) error  { return nil }
func (c channelConn) SetWriteDeadline(t time.Time) error { return nil }

func dialStreamLocal(sshClient *ssh.Client, socketPath string) (net.Conn, error) {
	if sshClient == nil {
		return nil, fmt.Errorf("ssh client is nil")
	}
	msg := streamLocalForwardMsg{SocketPath: socketPath, Reserved: "", Flags: 0}
	ch, reqs, err := sshClient.OpenChannel("direct-streamlocal@openssh.com", ssh.Marshal(msg))
	if err != nil {
		return nil, err
	}
	go ssh.DiscardRequests(reqs)
	return channelConn{ch}, nil
}

// Tunnel proxies local TCP to remote unix socket over SSH streamlocal.
type Tunnel struct {
	ln       net.Listener
	ssh      *ssh.Client
	socket   string
	wg       sync.WaitGroup
	once     sync.Once
	closeErr error
}

// StartTunnel listens on 127.0.0.1:0 and forwards to remote unix socket.
func StartTunnel(sshClient *ssh.Client, socketPath string) (*Tunnel, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	t := &Tunnel{ln: ln, ssh: sshClient, socket: socketPath}
	t.wg.Add(1)
	go t.acceptLoop()
	return t, nil
}

func (t *Tunnel) acceptLoop() {
	defer t.wg.Done()
	for {
		local, err := t.ln.Accept()
		if err != nil {
			return
		}
		go t.handle(local)
	}
}

func (t *Tunnel) handle(local net.Conn) {
	remote, err := dialStreamLocal(t.ssh, t.socket)
	if err != nil {
		_ = local.Close()
		return
	}
	go func() {
		_, _ = io.Copy(remote, local)
		_ = remote.Close()
		_ = local.Close()
	}()
	go func() {
		_, _ = io.Copy(local, remote)
		_ = remote.Close()
		_ = local.Close()
	}()
}

// Host returns tcp://127.0.0.1:PORT for docker client.
func (t *Tunnel) Host() string {
	if t == nil || t.ln == nil {
		return ""
	}
	return "tcp://" + t.ln.Addr().String()
}

// Close stops the tunnel listener.
func (t *Tunnel) Close() error {
	if t == nil {
		return nil
	}
	t.once.Do(func() {
		if t.ln != nil {
			t.closeErr = t.ln.Close()
		}
	})
	t.wg.Wait()
	return t.closeErr
}
