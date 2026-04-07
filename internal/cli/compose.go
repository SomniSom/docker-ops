package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/SomniSom/docker-ops/internal/locale"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// logsComposeExtras decides docker compose logs flags and whether the caller should use
// RunTTY instead of Run. For logs-tail (tailOnly true) it always returns --tail 200 and
// useRunTTY false. For streaming logs, if stdin is a terminal it returns -f and useRunTTY
// true so SIGINT can be forwarded; otherwise it uses --tail 200 because logs -f with a
// non-TTY stdin often exits immediately with no useful output.
func logsComposeExtras(tailOnly, stdinIsTerminal bool) (extra []string, useRunTTY bool) {
	if tailOnly {
		return []string{"--tail", "200"}, false
	}
	if stdinIsTerminal {
		return []string{"-f"}, true
	}
	return []string{"--tail", "200"}, false
}

// splitLeadingServiceArgs splits argv for dq logs / status: consecutive arguments that
// do not start with '-' are treated as compose service names; the remainder (typically
// docker compose flags like --tail) is returned as rest and appended after services
// when building the docker compose argv.
func splitLeadingServiceArgs(args []string) (services, rest []string) {
	i := 0
	for i < len(args) && !strings.HasPrefix(args[i], "-") {
		services = append(services, args[i])
		i++
	}
	return services, args[i:]
}

// newComposeCmds returns the compose-related subcommands bound to projectDir (pointer
// to the root persistent flag value).
func newComposeCmds(projectDir *string) []*cobra.Command {
	return []*cobra.Command{
		newBuildCmd(projectDir),
		newPullCmd(projectDir),
		newUpCmd(projectDir),
		newDownCmd(projectDir),
		newStopCmd(projectDir),
		newReupCmd(projectDir),
		newPsCmd(projectDir),
		newRestartCmd(projectDir),
		newExecCmd(projectDir),
		newStatusCmd(projectDir),
		newLogsCmd(projectDir, false),
		newLogsCmd(projectDir, true),
	}
}

// newBuildCmd creates "build": docker compose build --pull plus extra args.
func newBuildCmd(projectDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: locale.T("build.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newComposeSession(projectDir)
			if err != nil {
				return err
			}
			return s.Run(append([]string{"build", "--pull"}, args...)...)
		},
	}
}

// newPullCmd creates "pull": docker compose pull plus extra args.
func newPullCmd(projectDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "pull",
		Short: locale.T("pull.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newComposeSession(projectDir)
			if err != nil {
				return err
			}
			return s.Run(append([]string{"pull"}, args...)...)
		},
	}
}

// newUpCmd creates "up": checks app_config exists if set, docker compose up -d, then ps.
func newUpCmd(projectDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: locale.T("up.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newComposeSession(projectDir)
			if err != nil {
				return err
			}
			cfg := s.cfgRef()
			if err := checkAppConfigExists(s.projectRoot(), cfg.AppConfig); err != nil {
				return err
			}
			if err := s.Run(append([]string{"up", "-d"}, args...)...); err != nil {
				return err
			}
			return s.Run("ps")
		},
	}
}

// newDownCmd creates "down": docker compose down plus extra args.
func newDownCmd(projectDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: locale.T("down.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newComposeSession(projectDir)
			if err != nil {
				return err
			}
			return s.Run(append([]string{"down"}, args...)...)
		},
	}
}

// newStopCmd creates "stop": docker compose stop (containers stay, unlike down), then ps.
func newStopCmd(projectDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "stop [service...]",
		Short: locale.T("stop.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newComposeSession(projectDir)
			if err != nil {
				return err
			}
			if err := s.Run(append([]string{"stop"}, args...)...); err != nil {
				return err
			}
			return s.Run("ps")
		},
	}
}

// newReupCmd creates "reup": build --pull, up -d, then ps (with app_config check like up).
func newReupCmd(projectDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "reup",
		Short: locale.T("reup.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newComposeSession(projectDir)
			if err != nil {
				return err
			}
			cfg := s.cfgRef()
			if err := checkAppConfigExists(s.projectRoot(), cfg.AppConfig); err != nil {
				return err
			}
			if err := s.Run(append([]string{"build", "--pull"}, args...)...); err != nil {
				return err
			}
			if err := s.Run(append([]string{"up", "-d"}, args...)...); err != nil {
				return err
			}
			return s.Run("ps")
		},
	}
}

// newPsCmd creates "ps": docker compose ps plus extra args.
func newPsCmd(projectDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "ps",
		Short: locale.T("ps.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newComposeSession(projectDir)
			if err != nil {
				return err
			}
			return s.Run(append([]string{"ps"}, args...)...)
		},
	}
}

// newRestartCmd creates "restart": service from first arg or compose_service in config,
// then docker compose ps.
func newRestartCmd(projectDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: locale.T("restart.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newComposeSession(projectDir)
			if err != nil {
				return err
			}
			cfg := s.cfgRef()
			svc := cfg.ComposeService
			if len(args) > 0 {
				svc = args[0]
				args = args[1:]
			}
			if svc == "" {
				return fmt.Errorf("%s", locale.T("restart.err"))
			}
			all := append([]string{"restart", svc}, args...)
			if err := s.Run(all...); err != nil {
				return err
			}
			return s.Run("ps")
		},
	}
}

// newExecCmd creates "exec": uses RunExecTTY with -it when stdin is a TTY and -T not set;
// otherwise compose exec -T.
func newExecCmd(projectDir *string) *cobra.Command {
	var noTTY bool
	c := &cobra.Command{
		Use:   "exec SERVICE COMMAND [ARG...]",
		Short: locale.T("exec.short"),
		Long:  locale.T("exec.long"),
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newComposeSession(projectDir)
			if err != nil {
				return err
			}
			service := args[0]
			cmdArgs := args[1:]
			stdinTTY := term.IsTerminal(int(os.Stdin.Fd())) && !noTTY

			var full []string
			if stdinTTY {
				full = append([]string{"exec", "-it", service}, cmdArgs...)
				return s.RunExecTTY(full...)
			}
			full = append([]string{"exec", "-T", service}, cmdArgs...)
			return s.Run(full...)
		},
	}
	c.Flags().BoolVarP(&noTTY, "no-tty", "T", false, locale.T("exec.flag.no_tty"))
	return c
}

// newStatusCmd creates "status": prints ps -a then a fixed --tail 80 logs slice for
// optional service names (splitLeadingServiceArgs).
func newStatusCmd(projectDir *string) *cobra.Command {
	return &cobra.Command{
		Use:   "status [service...]",
		Short: locale.T("status.short"),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newComposeSession(projectDir)
			if err != nil {
				return err
			}
			cfg := s.cfgRef()
			fmt.Printf(locale.T("status.header.ps"), cfg.ComposeProjectName)
			if err := s.Run("ps", "-a"); err != nil {
				return err
			}
			services, rest := splitLeadingServiceArgs(args)
			label := locale.T("status.label.all")
			if len(services) > 0 {
				label = strings.Join(services, ", ")
			}
			fmt.Printf(locale.T("status.header.logs"), label)
			logArgs := append([]string{"logs", "--tail", "80", "--timestamps"}, services...)
			logArgs = append(logArgs, rest...)
			if err := s.Run(logArgs...); err != nil {
				fmt.Fprintf(os.Stderr, locale.T("status.err.logs"), err)
			}
			return nil
		},
	}
}

// newLogsCmd creates "logs" (tailOnly false) or "logs-tail" (tailOnly true). Follow mode
// uses RunTTY when stdin is a terminal; treats user interrupt exit codes 130/141 as success
// for local exec and SSH.
func newLogsCmd(projectDir *string, tailOnly bool) *cobra.Command {
	use := locale.T("logs.use.follow")
	short := locale.T("logs.short.follow")
	if tailOnly {
		use = locale.T("logs.use.tail")
		short = locale.T("logs.short.tail")
	}
	return &cobra.Command{
		Use:   use,
		Short: short,
		Long:  locale.T("logs.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := newComposeSession(projectDir)
			if err != nil {
				return err
			}
			services, rest := splitLeadingServiceArgs(args)
			extra, followTTY := logsComposeExtras(tailOnly, term.IsTerminal(int(os.Stdin.Fd())))

			base := append([]string{"logs"}, extra...)
			base = append(base, "--timestamps")
			base = append(base, services...)
			base = append(base, rest...)
			if !followTTY {
				return s.Run(base...)
			}
			err = s.RunTTY(base...)
			var sex *ssh.ExitError
			if errors.As(err, &sex) {
				if c := sex.ExitStatus(); c == 130 || c == 141 {
					return nil
				}
			}
			var ex *exec.ExitError
			if errors.As(err, &ex) {
				if code := ex.ExitCode(); code == 130 || code == 141 {
					return nil
				}
			}
			return err
		},
	}
}
