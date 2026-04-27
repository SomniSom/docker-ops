package locale

// catalogEn is the default (English) message catalog.
var catalogEn = map[string]string{
	"root.short": "Docker Compose locally and over SSH",
	"root.long":  "%s (%s) — manage Docker Compose projects. See readme.md for the full specification.",

	"flag.lang":         "UI language: en, ru, or auto (detect from LANG / LC_MESSAGES / DQ_LANG)",
	"flag.project_root": "project root (default: $DQ_PROJECT_ROOT or cwd)",

	"version.short": "Print version information",

	"validate.short": "Validate docker-ops.yaml / docker-ops.yml syntax and fields",
	"validate.long": `Parses the config file in the project root and reports clear errors for invalid YAML
(common mistake: a space before a top-level key like deploy_include). Also checks deploy_mode
and that app_config exists on disk if set.`,
	"validate.err.no_docker_ops": "no docker-ops.yaml or docker-ops.yml in %s",
	"validate.note.remote":       "\ndq: note: remote_ssh and remote_path are set — compose commands run over SSH. To force local Docker only: DOCKER_OPS_USE_REMOTE=0 or use_remote: false in YAML.\n",

	"configcheck.short":        "Verify app config file exists when app_config is set (readme §5.3)",
	"configcheck.info.not_set": "dq: app_config not set — nothing to check",
	"msg.ok_path":              "ok: %s",

	"deploy.short": "Deploy the project to the remote host over SSH (readme §5.3)",
	"deploy.long": `Sync files to the server over SFTP (source mode) or deliver compose + image (artifacts mode), then run remote docker compose (readme §5.3).

Source (default): mirror project tree with excludes (size/mtime), deploy_include, app_config, then remote reup.

Artifacts: requires docker-compose.image.yml. Single image: deploy_image. Multiple built services: deploy_images (map service name → image tag) and dq gen-image-compose --all-built with matching deploy_images; each service is docker built from its compose build context. Optional docker build/push or save|ssh|load; then remote pull+up or up. Flag --build runs local docker build for every image before save/load (or push) even if deploy_push is unset. With deploy_build_remote: true and a build (--build or deploy_push), the project tree is mirrored over SFTP and docker build (and registry push when not save/load) runs on the remote host instead of locally.`,
	"deploy.err.needs_remote": "deploy needs remote_ssh and remote_path (docker-ops.yml or dq.env); run 'dq validate' to check the file",
	"deploy.flag.build":       "artifacts: docker build -t deploy_image, then save/load or push as configured",

	"build.short":   "docker compose build --pull",
	"pull.short":    "docker compose pull",
	"up.short":      "docker compose up -d",
	"down.short":    "docker compose down",
	"stop.short":    "docker compose stop (containers kept; optional service names)",
	"reup.short":    "docker compose build --pull && up -d",
	"ps.short":      "docker compose ps",
	"restart.short": "docker compose restart <service>",
	"restart.err":   "compose_service is empty and no service argument given",

	"exec.short": "docker compose exec (interactive TTY when stdin is a terminal)",
	"exec.long": `Run a command in a running service container. When stdin is a TTY, uses docker compose exec -it (e.g. dq exec audio-bot bash).

Ctrl+C is forwarded to the container (interrupt current line). Ctrl+D (EOF) exits the shell. Without a TTY, uses exec -T. For a TTY but a non-interactive command, pass -T / --no-tty.`,
	"exec.flag.no_tty": "docker compose exec -T (disable pseudo-TTY)",

	"status.short":          "docker compose ps -a and recent logs (all services by default)",
	"status.header.ps":      "=== docker compose ps -a (project %s) ===\n",
	"status.header.logs":    "\n=== last 80 log lines (%s) ===\n",
	"status.err.logs":       "dq: logs: %v\n",
	"status.label.all":      "all services",

	"logs.use.follow":    "logs [service...]",
	"logs.use.tail":      "logs-tail [service...]",
	"logs.short.follow":  "docker compose logs -f (all services unless service names are given)",
	"logs.short.tail":    "docker compose logs --tail 200 (all services unless service names are given)",
	"logs.long": `Stream or print compose logs. With no service names, docker compose logs all services.

If stdin is not a terminal (e.g. remote run from an IDE), follow mode uses --tail 200 instead of -f so logs are not empty.

In follow mode from a terminal: Ctrl+C stops streaming — dq kills the local docker compose process tree, or closes the SSH session when using remote compose (SIGINT is not forwarded into docker compose logs). Ctrl+D does not exit log follow.

Optional service names are positional arguments before any flags, e.g.:
  dq logs
  dq logs parser
  dq logs parser worker --tail 50`,

	"env.short": "Print a docker-ops.yaml template (readme §5.4)",
	"env.err":   "refusing to overwrite %s (use --force)",
	"env.wrote": "dq: wrote %s\n",

	"completion.short": "Generate shell completion script",
	"completion.long": `To load completions:

Bash:
  source <(dq completion bash)

Zsh:
  source <(dq completion zsh)

Fish:
  dq completion fish | source
`,

	"paths.err.app_config": "app config %s",
	"paths.err.app_dir":    "app config is a directory: %s",

	"err.remote_not_configured": "remote not configured",
	"err.remote_ssh":            "remote SSH is not configured",
	"err.config_required":       "config required",

	"load.err.read":   "read config",
	"load.err.config": "config %s",

	"validate.err.read_file": "read %s",
	"validate.deploy_mode":         "deploy_mode must be \"source\", \"artifacts\", or empty; got %q",
	"validate.deploy_build_remote_artifacts": "deploy_build_remote: true requires deploy_mode: artifacts",
	"validate.app_config_missing": "app_config points to missing file %q (resolved: %s)",
	"validate.err.dq_env": "dq.env",

	"envfile.expected_kv": "%s:%d: expected KEY=value",
	"envfile.empty_key":   "%s:%d: empty key",

	"yaml.invalid_intro":    "invalid docker-ops YAML in %s\n\n",
	"yaml.parser_said":      "Parser said: %s\n",
	"yaml.context_header":   "\nContext (parser line %d):\n",
	"yaml.past_eof":         "\n(Reported line %d is past end of file; showing last lines.)\n",
	"yaml.no_line":          "\nCould not map error to a line number; check indentation and colons in keys.\n",
	"yaml.common_issues":    "Common issues:\n",
	"yaml.issue.root_keys":  "  • Top-level keys (remote_ssh, deploy_include, …) must start at column 1 — no leading spaces.\n",
	"yaml.issue.spaces_lists": "  • Use spaces for indentation under lists (e.g. deploy_include), not tabs mixed with spaces.\n",
	"yaml.issue.colon_quote": "  • Values with ':' should be quoted.\n",
	"yaml.misindented_root":   "Detected mis-indented root key %q at line %d — root keys must start at column 0 (no leading spaces or tabs).",
	"yaml.hint.tabs":        "Hint: line %d uses tabs before %q. Prefer spaces only, or fix indentation.",
	"yaml.hint.spaces":      "Hint: %q looks indented with %d leading space(s). Top-level keys in docker-ops.yml must start at the beginning of the line (column 0). Remove the spaces before %q.",
	"yaml.hint.key_space":   "Hint: key %q may contain a space; YAML keys cannot have unquoted spaces before ':'.",
	"yaml.hint.top_level":   "Top-level keys",

	"compose.docker_path": "docker not found in PATH",
	"compose.v2_required": `Docker Compose V2 plugin is required (command: docker compose).
Standalone docker-compose (V1) is not supported.

%s

Docker said: %s`,
	"compose.version_prefix": "docker compose version",
	"compose.install_hint": `Install the Compose V2 plugin, for example:
  • Docs: https://docs.docker.com/compose/install/linux/
  • Debian/Ubuntu: sudo apt-get update && sudo apt-get install -y docker-compose-plugin
  • Then verify: docker compose version`,
	"compose.run_prefix": "docker compose",
	"compose.file_missing": "compose file %s",

	"template.header1": "# docker-ops — Docker Quick-ops (generated by dq env)",
	"template.header2": "# Save as docker-ops.yaml in project root (do not commit secrets; use dq.env)",

	"deploy.src.remote_mkdir": "remote mkdir",
	"deploy.src.sftp":         "sftp",

	"deploy.art.docker_build": "docker build",
	"deploy.art.docker_push":  "docker push",
	"deploy.art.err.image":    "deploy_mode=artifacts requires deploy_image or deploy_images (docker-ops.yml or dq.env)",
	"deploy.art.err.images_empty": "deploy_images has no non-empty image tags",
	"deploy.art.err.no_build_service": "deploy_images: service %q has no build: in %s — add build: or remove the entry",
	"deploy.art.err.no_images": "internal: no images to transfer",
	"deploy.art.err.read_compose": "read compose file %s",
	"deploy.art.err.compose":  "artifacts deploy requires %s in the project root",
	"deploy.art.upload":       "upload compose",
	"deploy.art.remote_up_sl": "==> remote: compose up (image loaded locally)\n",
	"deploy.art.remote_pull":  "==> remote: compose pull && up\n",
	"deploy.art.save_gzip":    "==> docker save (gzip) | ssh … docker load\n",
	"deploy.art.save_plain":   "==> docker save | ssh … docker load\n",
	"deploy.art.build":        "==> docker build -t %s\n",
	"deploy.art.build_svc":    "==> docker build -t %s (service %s)\n",
	"deploy.art.push":         "==> docker push %s\n",
	"deploy.art.mirror_remote": "==> mirror project tree for remote build (SFTP)\n",
	"deploy.art.build_remote":  "==> remote: docker build -t %s\n",
	"deploy.art.build_remote_svc": "==> remote: docker build -t %s (service %s)\n",
	"deploy.art.push_remote":   "==> remote: docker push %s\n",

	"deploy.inc.skip_abs":    "dq: deploy_include: skip absolute path %q\n",
	"deploy.inc.skip_unsafe": "dq: deploy_include: skip unsafe path %q\n",
	"deploy.inc.missing":     "dq: warning: deploy_include: missing %s\n",
	"deploy.inc.err": "deploy_include %s",

	"deploy.mirror.list":   "list remote",
	"deploy.mirror.remove": "remote remove %s",
	"deploy.mirror.mkdir":  "mkdir %s",
	"deploy.mirror.upload": "upload %s",
	"deploy.mirror.not_under": "%q not under %q",

	"deploy.appcfg.warn": "dq: warning: app_config file missing locally (%s); remote copy skipped\n",
	"deploy.appcfg.dir":  "app_config is a directory: %s",

	"ssh.err.no_credentials": "no SSH credentials: set ssh_identity to an unencrypted key, or use ssh-agent (encrypted keys)",
	"ssh.err.empty_remote":   "empty remote_ssh",
	"ssh.err.bad_remote":     "remote_ssh must be user@host (got %q)",
	"ssh.err.no_host":        "remote_ssh missing host after @",
	"ssh.err.known_hosts":    "known_hosts",
	"ssh.err.kh_file":        "known_hosts %s",
	"ssh.err.host_key_suffix": "refusing: host key changed or mismatch — check MITM",
	"ssh.err.dial":           "ssh dial %s",
	"ssh.err.remote":         "remote",
	"ssh.err.stdin":          "ssh client and stdin required",
	"ssh.err.request_pty":    "request pty",

	"genimg.short": "Generate docker-compose.image.yml from compose_file for artifacts deploy",
	"genimg.long": `Rewrites the app service from build: to image: (default compose_service). Sidecars that only declare image: are unchanged.

Fails if any other service still has build: unless --all-built is set. With --all-built, either every built service gets the same image expression (default ${DEPLOY_IMAGE}), or set deploy_images in docker-ops.yml to map each service name to a full image tag (literal image: lines).

Examples:
  dq gen-image-compose
  dq gen-image-compose --service app -o docker-compose.image.yml
  dq gen-image-compose --all-built`,
	"genimg.header": "# Generated by dq gen-image-compose — use DEPLOY_IMAGE on deploy (docker-ops.yml / dq.env).",
	"genimg.wrote":  "wrote %s",
	"genimg.err.read":       "read %s",
	"genimg.err.no_service": "set compose_service in docker-ops or pass --service",
	"genimg.err.transform":  "compose transform",
	"genimg.err.write":      "write %s",
	"genimg.flag.output":        "output path relative to project root",
	"genimg.flag.compose_file":  "input compose file relative to project root (default: compose_file from config)",
	"genimg.flag.service":       "service to convert (default: compose_service)",
	"genimg.flag.image_expr":    "value for image: line (default ${DEPLOY_IMAGE})",
	"genimg.flag.all_built":     "convert every service with build:; use deploy_images in config for per-service tags",

	"man.short": "Open manual page (groff) for dq or a subcommand",
	"man.long": `Generates a man page from Cobra help and opens it with man -l (requires man-db / groff).

Examples:
  dq man
  dq man deploy
  dq man gen-image-compose

To install pages system-wide, run "make install-man" (see Makefile).`,
	"man.err.gen":     "generate man page",
	"man.hint.no_man": "\n(man not found in PATH; raw troff source printed above — pipe through groff -man -T utf8 | less if needed)",
}
