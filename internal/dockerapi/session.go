package dockerapi

import (
	"context"
)

// Pair holds local and optional remote docker clients for a deploy session.
type Pair struct {
	Local  *Client
	Remote *RemoteSession
}

// OpenPair opens local client (when needed) and remote session via SSH tunnel.
func OpenPair(ctx context.Context, needLocal bool, remote *RemoteSession) (*Pair, error) {
	p := &Pair{Remote: remote}
	if needLocal {
		cli, err := NewLocal(ctx)
		if err != nil {
			return nil, err
		}
		p.Local = cli
	}
	return p, nil
}

// Close closes all clients in the pair.
func (p *Pair) Close() {
	if p == nil {
		return
	}
	if p.Local != nil {
		_ = p.Local.Close()
	}
	if p.Remote != nil {
		_ = p.Remote.Close()
	}
}
