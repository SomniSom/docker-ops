package deploy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SomniSom/docker-ops/internal/composeimage"
	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/locale"
	"github.com/SomniSom/docker-ops/internal/sshexec"
	"golang.org/x/crypto/ssh"
)

// sortedDeployImageKeys returns service names with non-empty image refs, sorted.
func sortedDeployImageKeys(m map[string]string) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k, v := range m {
		if strings.TrimSpace(v) == "" {
			continue
		}
		if strings.TrimSpace(k) == "" {
			continue
		}
		out = append(out, strings.TrimSpace(k))
	}
	sort.Strings(out)
	return out
}

func runArtifactBuilds(projectRoot string, cfg *config.Config, opts RunOpts, baseCompose []byte) ([]string, error) {
	projectRoot = filepath.Clean(projectRoot)
	multi := cfg.DeployImages
	needBuild := EffectiveDeployPush(cfg) || opts.Build

	if len(multi) == 0 {
		img := strings.TrimSpace(cfg.DeployImage)
		if img == "" {
			return nil, fmt.Errorf("%s", locale.T("deploy.art.err.image"))
		}
		if !needBuild {
			return []string{img}, nil
		}
		fmt.Fprint(os.Stderr, locale.Tf("deploy.art.build", img))
		cmd := exec.Command("docker", "build", "-t", img, ".")
		cmd.Dir = projectRoot
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("%s: %w", locale.T("deploy.art.docker_build"), err)
		}
		return []string{img}, nil
	}

	ctxs, err := composeimage.ServiceBuildDirs(baseCompose, projectRoot)
	if err != nil {
		return nil, err
	}
	keys := sortedDeployImageKeys(multi)
	if len(keys) == 0 {
		return nil, fmt.Errorf("%s", locale.T("deploy.art.err.images_empty"))
	}
	var refs []string
	for _, svc := range keys {
		imgRef := strings.TrimSpace(multi[svc])
		ctxDir, ok := ctxs[svc]
		if !ok {
			return nil, fmt.Errorf("%s", locale.Tf("deploy.art.err.no_build_service", svc, cfg.ComposeFile))
		}
		if !needBuild {
			refs = append(refs, imgRef)
			continue
		}
		fmt.Fprint(os.Stderr, locale.Tf("deploy.art.build_svc", imgRef, svc))
		cmd := exec.Command("docker", "build", "-t", imgRef, ".")
		cmd.Dir = ctxDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("%s: %w", locale.T("deploy.art.docker_build"), err)
		}
		refs = append(refs, imgRef)
	}
	return refs, nil
}

func dockerSaveLoadMulti(client *ssh.Client, projectRoot string, images []string, compress bool) error {
	if len(images) == 0 {
		return fmt.Errorf("%s", locale.T("deploy.art.err.no_images"))
	}
	if len(images) == 1 {
		return dockerSaveLoad(client, projectRoot, images[0], compress)
	}
	if compress {
		fmt.Fprint(os.Stderr, locale.T("deploy.art.save_gzip"))
		args := append([]string{"save"}, images...)
		inner := fmt.Sprintf("docker %s | gzip -c", sshexec.QuoteArgs(args))
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
	args := append([]string{"save"}, images...)
	cmd := exec.Command("docker", args...)
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

func dockerPushMulti(projectRoot string, images []string) error {
	for _, img := range images {
		fmt.Fprint(os.Stderr, locale.Tf("deploy.art.push", img))
		cmd := exec.Command("docker", "push", img)
		cmd.Dir = projectRoot
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", locale.T("deploy.art.docker_push"), err)
		}
	}
	return nil
}
