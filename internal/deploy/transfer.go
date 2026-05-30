package deploy

import (
	"context"
	"fmt"
	"os"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/deploy/imagecmp"
	"github.com/SomniSom/docker-ops/internal/deploy/layersync"
	"github.com/SomniSom/docker-ops/internal/dockerapi"
	"github.com/SomniSom/docker-ops/internal/locale"
	"golang.org/x/crypto/ssh"
)

// transferOpts controls optional API-based image transfer.
type transferOpts struct {
	needLocal     bool
	useSaveLoad   bool
	layerSync     bool
	skipUnchanged bool
	tryAPI        bool
}

func transferArtifactsImages(ctx context.Context, sshClient *ssh.Client, cfg *config.Config, projectRoot string, imageRefs []string, opts transferOpts) (*dockerapi.RemoteSession, error) {
	if len(imageRefs) == 0 {
		return nil, fmt.Errorf("%s", locale.T("deploy.art.err.no_images"))
	}
	var remoteSess *dockerapi.RemoteSession
	if opts.tryAPI {
		s, err := dockerapi.TryOpenRemote(ctx, sshClient, cfg)
		if err == nil {
			remoteSess = s
		}
	}
	if remoteSess == nil {
		if opts.useSaveLoad {
			if err := dockerSaveLoadMulti(sshClient, projectRoot, imageRefs, EffectiveSaveCompress(cfg)); err != nil {
				return nil, err
			}
		} else {
			if err := dockerPushMulti(projectRoot, imageRefs); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}
	var localCLI *dockerapi.Client
	if opts.needLocal && opts.useSaveLoad {
		var err error
		localCLI, err = dockerapi.NewLocal(ctx)
		if err != nil {
			_ = remoteSess.Close()
			return nil, dockerSaveLoadMultiFallback(sshClient, projectRoot, imageRefs, EffectiveSaveCompress(cfg), err)
		}
		defer localCLI.Close()
	}
	refs := imageRefs
	if opts.skipUnchanged && localCLI != nil {
		need, skipped, err := imagecmp.FilterUnchanged(ctx, localCLI, remoteSess.Client, imageRefs, true)
		if err == nil {
			for _, s := range skipped {
				fmt.Fprint(os.Stderr, locale.Tf("deploy.skip_unchanged", s))
			}
			refs = need
		}
	}
	if len(refs) == 0 {
		return remoteSess, nil
	}
	if opts.useSaveLoad && localCLI != nil {
		for _, ref := range refs {
			if err := layersync.SyncImage(ctx, localCLI, remoteSess.Client, ref, opts.layerSync); err != nil {
				_ = remoteSess.Close()
				return nil, dockerSaveLoadMultiFallback(sshClient, projectRoot, refs, EffectiveSaveCompress(cfg), err)
			}
		}
		return remoteSess, nil
	}
	if !opts.useSaveLoad {
		if err := dockerPushMulti(projectRoot, refs); err != nil {
			_ = remoteSess.Close()
			return nil, err
		}
	}
	return remoteSess, nil
}

func dockerSaveLoadMultiFallback(sshClient *ssh.Client, projectRoot string, refs []string, compress bool, cause error) error {
	fmt.Fprint(os.Stderr, locale.Tf("deploy.api.transfer_fallback", cause))
	return dockerSaveLoadMulti(sshClient, projectRoot, refs, compress)
}

func buildTransferOpts(cfg *config.Config, opts RunOpts, useSL, needLocalBuild bool) transferOpts {
	tryAPI := config.UsesDeployAPI(cfg) || config.EffectiveDeploySkipUnchanged(cfg) || config.EffectiveDeployLayerSync(cfg)
	return transferOpts{
		needLocal:     needLocalBuild,
		useSaveLoad:   useSL,
		layerSync:     config.EffectiveDeployLayerSync(cfg),
		skipUnchanged: config.EffectiveDeploySkipUnchanged(cfg),
		tryAPI:        tryAPI,
	}
}
