package imagecmp

import (
	"context"
	"fmt"

	"github.com/SomniSom/docker-ops/internal/dockerapi"
)

// Info holds resolved image identity and layer chain.
type Info struct {
	ID     string
	Layers []string
}

// ResolveImageRef inspects ref and returns image ID and RootFS layers.
func ResolveImageRef(ctx context.Context, c *dockerapi.Client, ref string) (Info, error) {
	if c == nil || c.Moby() == nil {
		return Info{}, fmt.Errorf("docker client is nil")
	}
	ii, _, err := c.Moby().ImageInspectWithRaw(ctx, ref)
	if err != nil {
		return Info{}, err
	}
	return Info{ID: ii.ID, Layers: append([]string(nil), ii.RootFS.Layers...)}, nil
}

// SameImage reports whether local and remote have the same image ID for ref.
func SameImage(ctx context.Context, local, remote *dockerapi.Client, ref string) (bool, error) {
	if local == nil || remote == nil {
		return false, fmt.Errorf("local and remote clients required")
	}
	l, err := ResolveImageRef(ctx, local, ref)
	if err != nil {
		return false, err
	}
	r, err := ResolveImageRef(ctx, remote, ref)
	if err != nil {
		return false, err
	}
	if l.ID == "" || r.ID == "" {
		return false, nil
	}
	return l.ID == r.ID, nil
}

// FilterUnchanged returns refs that differ between local and remote.
func FilterUnchanged(ctx context.Context, local, remote *dockerapi.Client, refs []string, skipUnchanged bool) (needTransfer []string, skipped []string, err error) {
	if !skipUnchanged {
		return append([]string(nil), refs...), nil, nil
	}
	for _, ref := range refs {
		same, cmpErr := SameImage(ctx, local, remote, ref)
		if cmpErr != nil {
			needTransfer = append(needTransfer, ref)
			continue
		}
		if same {
			skipped = append(skipped, ref)
		} else {
			needTransfer = append(needTransfer, ref)
		}
	}
	return needTransfer, skipped, nil
}
