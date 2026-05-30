package deploy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/dockerapi"
	"github.com/SomniSom/docker-ops/internal/locale"
	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"golang.org/x/crypto/ssh"
)

// ApplyArtifactsStack creates/updates containers from docker-compose.image.yml via Docker API.
func ApplyArtifactsStack(ctx context.Context, remote *dockerapi.Client, cfg *config.Config, composeFile, projectRoot string) error {
	if remote == nil || cfg == nil {
		return fmt.Errorf("apply: missing client or config")
	}
	b, err := os.ReadFile(composeFile)
	if err != nil {
		return err
	}
	project, err := loader.LoadWithContext(ctx, types.ConfigDetails{
		WorkingDir: projectRoot,
		ConfigFiles: []types.ConfigFile{{
			Filename: filepath.Base(composeFile),
			Content:  b,
		}},
	}, func(o *loader.Options) {
		o.SetProjectName(cfg.ComposeProjectName, false)
	})
	if err != nil {
		return fmt.Errorf("%s: %w", locale.T("deploy.api.err.parse_compose"), err)
	}
	applyImageRefs(cfg, project)
	for _, name := range sortedServiceNames(project) {
		svc := project.Services[name]
		if err := applyService(ctx, remote, cfg, project, name, svc); err != nil {
			return err
		}
	}
	return nil
}

func sortedServiceNames(p *types.Project) []string {
	out := make([]string, 0, len(p.Services))
	for n := range p.Services {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func applyImageRefs(cfg *config.Config, p *types.Project) {
	if len(cfg.DeployImages) > 0 {
		for name, ref := range cfg.DeployImages {
			if svc, ok := p.Services[name]; ok {
				svc.Image = strings.TrimSpace(ref)
				p.Services[name] = svc
			}
		}
		return
	}
	di := strings.TrimSpace(cfg.DeployImage)
	if di == "" {
		return
	}
	for name, svc := range p.Services {
		if strings.Contains(svc.Image, "${DEPLOY_IMAGE}") || strings.Contains(svc.Image, "$DEPLOY_IMAGE") {
			svc.Image = di
			p.Services[name] = svc
		}
	}
}

func applyService(ctx context.Context, remote *dockerapi.Client, cfg *config.Config, project *types.Project, name string, svc types.ServiceConfig) error {
	if strings.TrimSpace(svc.Image) == "" {
		return fmt.Errorf("%s", locale.Tf("deploy.api.err.no_image", name))
	}
	projectName := cfg.ComposeProjectName
	labels := map[string]string{
		"com.docker.compose.project": projectName,
		"com.docker.compose.service": name,
		"com.docker.compose.oneoff":  "False",
	}
	for k, v := range svc.Labels {
		labels[k] = v
	}
	hostConfig, err := serviceHostConfig(svc, project)
	if err != nil {
		return err
	}
	containerConfig := &container.Config{
		Image:  svc.Image,
		Env:    mappingToEnv(svc.Environment),
		Labels: labels,
	}
	networking := serviceNetworking(svc)
	existing, err := findComposeContainer(ctx, remote, projectName, name)
	if err != nil {
		return err
	}
	if existing != "" && containerUnchanged(ctx, remote, existing, containerConfig, hostConfig) {
		fmt.Fprint(os.Stderr, locale.Tf("deploy.api.skip_unchanged", name))
		return nil
	}
	if existing != "" {
		timeout := 10
		_ = remote.Moby().ContainerStop(ctx, existing, container.StopOptions{Timeout: &timeout})
		_ = remote.Moby().ContainerRemove(ctx, existing, container.RemoveOptions{Force: true})
	}
	createName := name
	if svc.ContainerName != "" {
		createName = svc.ContainerName
	}
	resp, err := remote.Moby().ContainerCreate(ctx, containerConfig, hostConfig, networking, nil, createName)
	if err != nil {
		return fmt.Errorf("%s: %w", locale.Tf("deploy.api.err.create", name), err)
	}
	if err := remote.Moby().ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("%s: %w", locale.Tf("deploy.api.err.start", name), err)
	}
	fmt.Fprint(os.Stderr, locale.Tf("deploy.api.started", name))
	return nil
}

func mappingToEnv(m types.MappingWithEquals) []string {
	mp := m.ToMapping()
	out := make([]string, 0, len(mp))
	for k, v := range mp {
		out = append(out, k+"="+v)
	}
	sort.Strings(out)
	return out
}

func findComposeContainer(ctx context.Context, remote *dockerapi.Client, project, service string) (string, error) {
	args := filters.NewArgs(
		filters.Arg("label", "com.docker.compose.project="+project),
		filters.Arg("label", "com.docker.compose.service="+service),
	)
	list, err := remote.Moby().ContainerList(ctx, container.ListOptions{All: true, Filters: args})
	if err != nil {
		return "", err
	}
	if len(list) == 0 {
		return "", nil
	}
	return list[0].ID, nil
}

func containerUnchanged(ctx context.Context, remote *dockerapi.Client, id string, cfg *container.Config, host *container.HostConfig) bool {
	insp, err := remote.Moby().ContainerInspect(ctx, id)
	if err != nil {
		return false
	}
	img, _, err := remote.Moby().ImageInspectWithRaw(ctx, cfg.Image)
	if err != nil {
		return false
	}
	if insp.Image != img.ID {
		return false
	}
	if !envEqual(insp.Config.Env, cfg.Env) {
		return false
	}
	if !mountsEqual(insp.HostConfig.Binds, host.Binds) {
		return false
	}
	return true
}

func envEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[string]struct{}, len(a))
	for _, e := range a {
		set[e] = struct{}{}
	}
	for _, e := range b {
		if _, ok := set[e]; !ok {
			return false
		}
	}
	return true
}

func mountsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[string]struct{}, len(a))
	for _, m := range a {
		set[m] = struct{}{}
	}
	for _, m := range b {
		if _, ok := set[m]; !ok {
			return false
		}
	}
	return true
}

func serviceHostConfig(svc types.ServiceConfig, project *types.Project) (*container.HostConfig, error) {
	hc := &container.HostConfig{
		RestartPolicy: restartPolicyFromString(svc.Restart),
	}
	for _, p := range svc.Ports {
		hostIP := p.HostIP
		if hostIP == "" {
			hostIP = "0.0.0.0"
		}
		proto := p.Protocol
		if proto == "" {
			proto = "tcp"
		}
		if hc.PortBindings == nil {
			hc.PortBindings = nat.PortMap{}
		}
		portKey := nat.Port(fmt.Sprintf("%d/%s", p.Target, proto))
		hc.PortBindings[portKey] = []nat.PortBinding{{
			HostIP:   hostIP,
			HostPort: p.Published,
		}}
	}
	for _, v := range svc.Volumes {
		src := v.Source
		if src == "" {
			continue
		}
		if !filepath.IsAbs(src) && project != nil {
			src = filepath.Join(project.WorkingDir, src)
		}
		hc.Binds = append(hc.Binds, fmt.Sprintf("%s:%s", src, v.Target))
	}
	return hc, nil
}

func restartPolicyFromString(r string) container.RestartPolicy {
	switch strings.ToLower(strings.TrimSpace(r)) {
	case "always":
		return container.RestartPolicy{Name: container.RestartPolicyAlways}
	case "on-failure":
		return container.RestartPolicy{Name: container.RestartPolicyOnFailure}
	case "unless-stopped":
		return container.RestartPolicy{Name: container.RestartPolicyUnlessStopped}
	case "no", "":
		return container.RestartPolicy{Name: container.RestartPolicyDisabled}
	default:
		return container.RestartPolicy{Name: container.RestartPolicyMode(r)}
	}
}

func serviceNetworking(svc types.ServiceConfig) *network.NetworkingConfig {
	if len(svc.Networks) == 0 {
		return nil
	}
	cfg := &network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{}}
	for netName := range svc.Networks {
		cfg.EndpointsConfig[netName] = &network.EndpointSettings{}
	}
	return cfg
}

// RunArtifactsFinish runs API apply or compose fallback per deploy_engine.
func RunArtifactsFinish(ctx context.Context, sshClient *ssh.Client, cfg *config.Config, composeFileOverride string, skipPull, exportDeployImage bool, composeLocal, projectRoot string, remoteSess *dockerapi.RemoteSession) error {
	engine := config.EffectiveDeployEngine(cfg)
	if engine == config.DeployEngineCompose {
		return RunRemoteArtifactsFinish(sshClient, cfg, composeFileOverride, skipPull, exportDeployImage)
	}
	if remoteSess == nil || remoteSess.Client == nil {
		if engine == config.DeployEngineAPI {
			return fmt.Errorf("%s", locale.T("deploy.api.err.no_client"))
		}
		fmt.Fprint(os.Stderr, locale.T("deploy.api.fallback_compose"))
		return RunRemoteArtifactsFinish(sshClient, cfg, composeFileOverride, skipPull, exportDeployImage)
	}
	composePath := composeLocal
	if composePath == "" {
		composePath = filepath.Join(projectRoot, composeFileOverride)
	}
	if err := ApplyArtifactsStack(ctx, remoteSess.Client, cfg, composePath, projectRoot); err != nil {
		if engine == config.DeployEngineAPI {
			return err
		}
		fmt.Fprint(os.Stderr, locale.Tf("deploy.api.fallback_compose_err", err))
		return RunRemoteArtifactsFinish(sshClient, cfg, composeFileOverride, skipPull, exportDeployImage)
	}
	return nil
}
