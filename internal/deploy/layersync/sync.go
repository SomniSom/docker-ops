package layersync

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/SomniSom/docker-ops/internal/deploy/imagecmp"
	"github.com/SomniSom/docker-ops/internal/dockerapi"
	"github.com/docker/docker/api/types/image"
)

// IndexLayerDigests returns set of layer diff IDs present on the daemon.
func IndexLayerDigests(ctx context.Context, c *dockerapi.Client) (map[string]struct{}, error) {
	if c == nil || c.Moby() == nil {
		return nil, fmt.Errorf("docker client is nil")
	}
	list, err := c.Moby().ImageList(ctx, image.ListOptions{})
	if err != nil {
		return nil, err
	}
	set := make(map[string]struct{})
	for _, img := range list {
		ii, _, err := c.Moby().ImageInspectWithRaw(ctx, img.ID)
		if err != nil {
			continue
		}
		for _, layer := range ii.RootFS.Layers {
			set[layer] = struct{}{}
		}
	}
	return set, nil
}

// MissingLayers returns ordered layer IDs from local image not in remote index.
func MissingLayers(ctx context.Context, local, remote *dockerapi.Client, ref string) (missing []string, localInfo imagecmp.Info, err error) {
	localInfo, err = imagecmp.ResolveImageRef(ctx, local, ref)
	if err != nil {
		return nil, localInfo, err
	}
	remoteIdx, err := IndexLayerDigests(ctx, remote)
	if err != nil {
		return nil, localInfo, err
	}
	for _, layer := range localInfo.Layers {
		if _, ok := remoteIdx[layer]; !ok {
			missing = append(missing, layer)
		}
	}
	return missing, localInfo, nil
}

// Import loads image tar stream on remote daemon.
func Import(ctx context.Context, remote *dockerapi.Client, r io.Reader) error {
	if remote == nil || remote.Moby() == nil {
		return fmt.Errorf("remote client is nil")
	}
	resp, err := remote.Moby().ImageLoad(ctx, r, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

// ExportPartial builds a docker-load tar containing only missing layer entries plus manifest/config.
func ExportPartial(ctx context.Context, local *dockerapi.Client, ref string, missing []string) (io.ReadCloser, error) {
	if len(missing) == 0 {
		return nil, fmt.Errorf("no missing layers")
	}
	missingSet := make(map[string]struct{}, len(missing))
	for _, m := range missing {
		missingSet[m] = struct{}{}
	}
	saveResp, err := local.Moby().ImageSave(ctx, []string{ref})
	if err != nil {
		return nil, err
	}
	defer saveResp.Close()
	raw, err := io.ReadAll(saveResp)
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(bytes.NewReader(raw))
	var manifestJSON []byte
	layerEntries := map[string][]byte{}
	configName := ""
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		name := hdr.Name
		switch {
		case name == "manifest.json":
			manifestJSON = data
		case name == "config.json":
			configName = name
			layerEntries[name] = data
		case !strings.Contains(name, "/") && strings.HasSuffix(name, ".json"):
			configName = name
			layerEntries[name] = data
		default:
			for layerID := range missingSet {
				if name == layerID+"/layer.tar" || name == layerID+"/VERSION" || name == layerID+"/json" {
					layerEntries[name] = data
				}
			}
		}
	}
	if len(manifestJSON) == 0 {
		return nil, fmt.Errorf("manifest.json not found in image save")
	}
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	writeEntry := func(name string, data []byte) error {
		if len(data) == 0 {
			return nil
		}
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0644, Size: int64(len(data))}); err != nil {
			return err
		}
		_, err := tw.Write(data)
		return err
	}
	if err := writeEntry("manifest.json", manifestJSON); err != nil {
		return nil, err
	}
	if configName != "" {
		if err := writeEntry(configName, layerEntries[configName]); err != nil {
			return nil, err
		}
	}
	names := slices.Sorted(maps.Keys(layerEntries))
	for _, name := range names {
		if name == configName {
			continue
		}
		if err := writeEntry(name, layerEntries[name]); err != nil {
			return nil, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

// SyncImage transfers ref to remote using partial layers when possible.
func SyncImage(ctx context.Context, local, remote *dockerapi.Client, ref string, layerSync bool) error {
	if !layerSync {
		return fullSaveLoad(ctx, local, remote, ref)
	}
	missing, info, err := MissingLayers(ctx, local, remote, ref)
	if err != nil {
		return fullSaveLoad(ctx, local, remote, ref)
	}
	if len(missing) == 0 {
		return tagIfNeeded(ctx, remote, ref, info.ID)
	}
	if len(missing) == len(info.Layers) {
		return fullSaveLoad(ctx, local, remote, ref)
	}
	r, err := ExportPartial(ctx, local, ref, missing)
	if err != nil {
		return fullSaveLoad(ctx, local, remote, ref)
	}
	defer r.Close()
	if err := Import(ctx, remote, r); err != nil {
		return fullSaveLoad(ctx, local, remote, ref)
	}
	return tagIfNeeded(ctx, remote, ref, info.ID)
}

func fullSaveLoad(ctx context.Context, local, remote *dockerapi.Client, ref string) error {
	saveResp, err := local.Moby().ImageSave(ctx, []string{ref})
	if err != nil {
		return err
	}
	defer saveResp.Close()
	return Import(ctx, remote, saveResp)
}

func tagIfNeeded(ctx context.Context, remote *dockerapi.Client, ref, id string) error {
	ii, _, err := remote.Moby().ImageInspectWithRaw(ctx, ref)
	if err == nil && ii.ID == id {
		return nil
	}
	return remote.Moby().ImageTag(ctx, id, ref)
}
