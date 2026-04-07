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

// logsComposeExtras returns flags after "logs" and whether to use an interactive TTY session (for Ctrl+C on follow).
// When stdin is not a terminal, docker compose logs -f exits immediately with little or no output; use --tail instead.
func logsComposeExtras(tailOnly, stdinIsTerminal bool) (extra []string, useRunTTY bool) {
	if tailOnly {
		return []string{"--tail", "200"}, false
	}
	if stdinIsTerminal {
		return []string{"-f"}, true
	}
	return []string{"--tail", "200"}, false
}

// splitLeadingServiceArgs treats initial args without a leading '-' as compose service names; the rest are passed to docker compose (flags, etc.).
func splitLeadingServiceArgs(args []string) (services, rest []string) {
	i := 0
	for i < len(args) && !strings.HasPrefix(args[i], "-") {
		services = append(services, args[i])
		i++
	}
	return services, args[i:]
}

func newComposeCmds(projectDir *string) []*cobra.Command {
	return []*cobra.Command{
		newBuildCmd(projectDir),
		newPullCmd(projectDir),
		newUpCmd(projectDir),
		newDownCmd(projectDir),
		newReupCmd(projectDir),
		newPsCmd(projectDir),
		newRestartCmd(projectDir),
		newExecCmd(projectDir),
		newStatusCmd(projectDir),
		newLogsCmd(projectDir, false),
		newLogsCmd(projectDir, true),
	}
}

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
				switch code := ex.ExitCode(); code {
				case 130, 141, 137, 143: // interrupt, pipe, SIGKILL, SIGTERM — expected when stopping logs -f
					return nil
				}
			}
			return err
		},
	}
}
