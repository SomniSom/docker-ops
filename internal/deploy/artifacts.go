package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/locale"
	"github.com/SomniSom/docker-ops/internal/sshexec"
	"golang.org/x/crypto/ssh"
)

const artifactsComposeFile = "docker-compose.image.yml"

// RunArtifacts builds/pushes or save-loads the image(s), syncs compose + config + includes, then remote pull+up or up (readme §5.3).
func RunArtifacts(projectRoot string, cfg *config.Config, opts RunOpts) error {
	if cfg == nil || !cfg.RemoteConfigured() {
		return fmt.Errorf("%s", locale.T("err.remote_not_configured"))
	}
	projectRoot = filepath.Clean(projectRoot)
	composeLocal := filepath.Join(projectRoot, artifactsComposeFile)
	if st, err := os.Stat(composeLocal); err != nil || st.IsDir() {
		return fmt.Errorf("%s", locale.Tf("deploy.art.err.compose", artifactsComposeFile))
	}

	baseComposePath := filepath.Join(projectRoot, cfg.ComposeFile)
	baseComposeBytes, err := os.ReadFile(baseComposePath)
	if err != nil {
		return fmt.Errorf("%s: %w", locale.Tf("deploy.art.err.read_compose", baseComposePath), err)
	}

	if DeployBuildRemote(cfg, opts) {
		return runArtifactsRemoteBuild(projectRoot, cfg, opts, baseComposeBytes, composeLocal)
	}

	imageRefs, err := runArtifactBuilds(projectRoot, cfg, opts, baseComposeBytes)
	if err != nil {
		return err
	}

	useSL := ArtifactsUseSaveLoad(cfg)

	client, err := sshexec.Dial(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	if useSL {
		if err := dockerSaveLoadMulti(client, projectRoot, imageRefs, EffectiveSaveCompress(cfg)); err != nil {
			return err
		}
	} else if !useSL && (EffectiveDeployPush(cfg) || opts.Build) {
		if err := dockerPushMulti(projectRoot, imageRefs); err != nil {
			return err
		}
	}

	rp := strings.TrimSpace(cfg.RemotePath)
	if err := sshexec.RunBash(client, "mkdir -p "+sshexec.ShellQuote(rp), false); err != nil {
		return fmt.Errorf("%s: %w", locale.T("deploy.src.remote_mkdir"), err)
	}

	c, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("%s: %w", locale.T("deploy.src.sftp"), err)
	}

	remCompose := remoteJoin(rp, artifactsComposeFile)
	if err := putLocalFile(c, composeLocal, remCompose); err != nil {
		_ = c.Close()
		return fmt.Errorf("%s: %w", locale.T("deploy.art.upload"), err)
	}
	if err := UploadAppConfig(c, projectRoot, rp, cfg); err != nil {
		_ = c.Close()
		return err
	}
	if err := SyncDeployIncludes(c, projectRoot, rp, cfg); err != nil {
		_ = c.Close()
		return err
	}
	if err := c.Close(); err != nil {
		return err
	}

	skipPull := useSL
	if skipPull {
		fmt.Fprint(os.Stderr, locale.T("deploy.art.remote_up_sl"))
	} else {
		fmt.Fprint(os.Stderr, locale.T("deploy.art.remote_pull"))
	}

	composeBytes, err := os.ReadFile(composeLocal)
	if err != nil {
		return err
	}
	exportDeployImage := strings.Contains(string(composeBytes), "${DEPLOY_IMAGE}")
	return RunRemoteArtifactsFinish(client, cfg, artifactsComposeFile, skipPull, exportDeployImage)
}

// runArtifactsRemoteBuild mirrors the project over SFTP, runs docker build on the remote host,
// optionally docker push (when not save/load), then remote compose up (readme deploy_build_remote).
func runArtifactsRemoteBuild(projectRoot string, cfg *config.Config, opts RunOpts, baseComposeBytes []byte, composeLocal string) error {
	rp := strings.TrimSpace(cfg.RemotePath)

	client, err := sshexec.Dial(cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := sshexec.RunBash(client, "mkdir -p "+sshexec.ShellQuote(rp), false); err != nil {
		return fmt.Errorf("%s: %w", locale.T("deploy.src.remote_mkdir"), err)
	}

	c, err := sftp.NewClient(client)
	if err != nil {
		return fmt.Errorf("%s: %w", locale.T("deploy.src.sftp"), err)
	}

	fmt.Fprint(os.Stderr, locale.T("deploy.art.mirror_remote"))
	patterns := MergeExcludePatterns(cfg)
	if err := MirrorProjectTree(c, projectRoot, rp, patterns); err != nil {
		_ = c.Close()
		return err
	}
	if err := UploadAppConfig(c, projectRoot, rp, cfg); err != nil {
		_ = c.Close()
		return err
	}
	if err := SyncDeployIncludes(c, projectRoot, rp, cfg); err != nil {
		_ = c.Close()
		return err
	}
	remCompose := remoteJoin(rp, artifactsComposeFile)
	if err := putLocalFile(c, composeLocal, remCompose); err != nil {
		_ = c.Close()
		return fmt.Errorf("%s: %w", locale.T("deploy.art.upload"), err)
	}
	if err := c.Close(); err != nil {
		return err
	}

	imageRefs, err := runRemoteArtifactBuilds(client, projectRoot, rp, cfg, opts, baseComposeBytes)
	if err != nil {
		return err
	}

	useSL := ArtifactsUseSaveLoad(cfg)
	if !useSL && (EffectiveDeployPush(cfg) || opts.Build) {
		if err := dockerPushMultiRemote(client, imageRefs); err != nil {
			return err
		}
	}

	skipPull := useSL
	if skipPull {
		fmt.Fprint(os.Stderr, locale.T("deploy.art.remote_up_sl"))
	} else {
		fmt.Fprint(os.Stderr, locale.T("deploy.art.remote_pull"))
	}

	composeBytes, err := os.ReadFile(composeLocal)
	if err != nil {
		return err
	}
	exportDeployImage := strings.Contains(string(composeBytes), "${DEPLOY_IMAGE}")
	return RunRemoteArtifactsFinish(client, cfg, artifactsComposeFile, skipPull, exportDeployImage)
}

func putLocalFile(c *sftp.Client, localPath, rem string) error {
	st, err := os.Stat(localPath)
	if err != nil {
		return err
	}
	parent := path.Dir(filepath.ToSlash(rem))
	if parent != "." && parent != "/" {
		if err := sftpMkdirAll(c, parent); err != nil {
			return err
		}
	}
	return uploadFile(c, localPath, rem, st)
}

func dockerSaveLoad(client *ssh.Client, projectRoot, image string, compress bool) error {
	if compress {
		fmt.Fprint(os.Stderr, locale.T("deploy.art.save_gzip"))
		inner := fmt.Sprintf("docker %s | gzip -c", sshexec.QuoteArgs([]string{"save", image}))
		cmd := exec.Command("bash", "-c", inner)
		cmd.Dir = projectRoot
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			return err
		}
		err = sshexec.RunRemotePipe(client, "gunzip -c | docker load", stdout)
		waitErr := cmd.Wait()
		if err != nil {
			return err
		}
		return waitErr
	}
	fmt.Fprint(os.Stderr, locale.T("deploy.art.save_plain"))
	cmd := exec.Command("docker", "save", image)
	cmd.Dir = projectRoot
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	err = sshexec.RunRemotePipe(client, "docker load", stdout)
	waitErr := cmd.Wait()
	if err != nil {
		return err
	}
	return waitErr
}
