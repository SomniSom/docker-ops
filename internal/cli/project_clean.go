package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SomniSom/docker-ops/internal/config"
	"github.com/SomniSom/docker-ops/internal/locale"
)

type cleanOpts struct {
	AllImages bool
	Volumes   bool
}

func runProjectClean(s *composeSession, opts cleanOpts) error {
	if s.cfg == nil || !s.cfg.RemoteConfigured() {
		return fmt.Errorf("%s", locale.T("clean.err.needs_remote"))
	}
	project := strings.TrimSpace(s.cfg.ComposeProjectName)
	if project == "" {
		return fmt.Errorf("%s", locale.T("clean.err.no_project"))
	}
	fmt.Fprint(os.Stderr, locale.Tf("clean.step.project", project))

	if err := s.dockerRun("container", "prune", "-f", "--filter",
		"label=com.docker.compose.project="+project); err != nil {
		return err
	}

	refs, err := projectImageRefs(s.cfg, s.localRoot)
	if err != nil {
		return err
	}
	candidates, err := s.projectImageCandidates(refs, opts.AllImages)
	if err != nil {
		return err
	}
	for _, ref := range candidates {
		inUse, err := s.dockerOutput("ps", "-q", "--filter", "ancestor="+ref)
		if err != nil {
			return err
		}
		if len(strings.TrimSpace(string(inUse))) > 0 {
			continue
		}
		fmt.Fprint(os.Stderr, locale.Tf("clean.step.rmi", ref))
		_ = s.dockerRun("rmi", "-f", ref)
	}

	if err := s.dockerRun("network", "prune", "-f", "--filter",
		"label=com.docker.compose.project="+project); err != nil {
		return err
	}

	if !opts.Volumes {
		return nil
	}
	return s.pruneProjectVolumes(project)
}

func projectImageRefs(cfg *config.Config, projectRoot string) ([]string, error) {
	set := map[string]struct{}{}
	add := func(ref string) {
		ref = strings.TrimSpace(ref)
		if ref != "" && !strings.HasPrefix(ref, "${") {
			set[ref] = struct{}{}
		}
	}
	if cfg != nil {
		add(cfg.DeployImage)
		for _, v := range cfg.DeployImages {
			add(v)
		}
	}
	fromCompose, err := imageRefsFromComposeFile(cfg, projectRoot)
	if err != nil {
		return nil, err
	}
	for _, ref := range fromCompose {
		add(ref)
	}
	out := make([]string, 0, len(set))
	for ref := range set {
		out = append(out, ref)
	}
	sort.Strings(out)
	return out, nil
}

func imageRefsFromComposeFile(cfg *config.Config, projectRoot string) ([]string, error) {
	if cfg == nil || projectRoot == "" {
		return nil, nil
	}
	cf := config.EffectiveComposeFile(cfg, projectRoot)
	path := filepath.Join(projectRoot, cf)
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}
	deployImage := strings.TrimSpace(cfg.DeployImage)
	var refs []string
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "image:") {
			continue
		}
		ref := strings.TrimSpace(strings.TrimPrefix(line, "image:"))
		ref = strings.Trim(ref, `"'`)
		if ref == "${DEPLOY_IMAGE}" || strings.HasPrefix(ref, "${DEPLOY_IMAGE:") {
			ref = deployImage
		}
		if ref != "" && !strings.HasPrefix(ref, "${") {
			refs = append(refs, ref)
		}
	}
	return refs, nil
}

func imageRepository(ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}
	if i := strings.Index(ref, "@"); i >= 0 {
		ref = ref[:i]
	}
	if i := strings.LastIndex(ref, ":"); i > 0 {
		suffix := ref[i+1:]
		if !strings.Contains(suffix, "/") {
			return ref[:i]
		}
	}
	return ref
}

func (s *composeSession) projectImageCandidates(refs []string, allImages bool) ([]string, error) {
	if !allImages {
		return refs, nil
	}
	repos := map[string]struct{}{}
	for _, ref := range refs {
		if repo := imageRepository(ref); repo != "" {
			repos[repo] = struct{}{}
		}
	}
	var candidates []string
	seen := map[string]struct{}{}
	for repo := range repos {
		out, err := s.dockerOutput("images", repo, "--format", "{{.Repository}}:{{.Tag}}")
		if err != nil {
			return nil, err
		}
		for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || line == "<none>:<none>" {
				continue
			}
			if _, ok := seen[line]; ok {
				continue
			}
			seen[line] = struct{}{}
			candidates = append(candidates, line)
		}
	}
	sort.Strings(candidates)
	return candidates, nil
}

func (s *composeSession) pruneProjectVolumes(project string) error {
	out, err := s.dockerOutput("volume", "ls", "-q", "--filter",
		"label=com.docker.compose.project="+project)
	if err != nil {
		return err
	}
	for _, vol := range strings.Fields(string(out)) {
		inUse, err := s.dockerOutput("ps", "-q", "--filter", "volume="+vol)
		if err != nil {
			return err
		}
		if len(strings.TrimSpace(string(inUse))) > 0 {
			continue
		}
		fmt.Fprint(os.Stderr, locale.Tf("clean.step.volume_rm", vol))
		_ = s.dockerRun("volume", "rm", "-f", vol)
	}
	return nil
}
