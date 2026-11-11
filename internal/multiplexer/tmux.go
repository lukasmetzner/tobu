package multiplexer

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/lukasmetzner/tobu/internal/command"
)

var tmuxNameReplacer = strings.NewReplacer(".", "_", ":", "_")

type Tmux struct {
	binary    string
	nameDepth int
}

func (t *Tmux) SessionName(dir string) string {
	return tmuxNameReplacer.Replace(sessionName(dir, t.nameDepth))
}

func (t *Tmux) Sessions(ctx context.Context) ([]string, error) {
	out, err := command.Output(ctx, t.binary, "list-sessions", "-F", "#{session_name}")
	if err != nil {
		if strings.Contains(err.Error(), "no server running") || strings.Contains(err.Error(), "error connecting to") {
			return nil, nil
		}
		return nil, err
	}
	return command.Lines(out), nil
}

func (t *Tmux) CurrentSession(ctx context.Context) (string, error) {
	if !insideTmux() {
		return "", nil
	}
	out, err := command.Output(ctx, t.binary, "display-message", "-p", "#{session_name}")
	return strings.TrimSpace(string(out)), err
}

func (t *Tmux) Open(_ context.Context, s Session) error {
	if !t.hasSession(s.Name) {
		if err := run(s.Dir, t.binary, tmuxNewSessionArgs(s)...); err != nil {
			return err
		}
		for _, c := range s.Commands[1:] {
			if err := run(s.Dir, t.binary, tmuxNewWindowArgs(s, c)...); err != nil {
				return err
			}
		}
	}
	return t.Attach(s.Name)
}

func (t *Tmux) Attach(name string) error {
	if !insideTmux() {
		return replaceProcess("", t.binary, "attach-session", "-t", tmuxTarget(name))
	}
	return run("", t.binary, "switch-client", "-t", tmuxTarget(name))
}

func (t *Tmux) hasSession(name string) bool {
	return exec.Command(t.binary, "has-session", "-t", tmuxTarget(name)).Run() == nil
}

func tmuxNewSessionArgs(s Session) []string {
	return append([]string{"new-session", "-d", "-s", s.Name, "-c", s.Dir, "--"}, s.Commands[0]...)
}

func tmuxNewWindowArgs(s Session, command []string) []string {
	return append([]string{"new-window", "-d", "-t", tmuxTarget(s.Name) + ":", "-c", s.Dir, "--"}, command...)
}

func tmuxTarget(name string) string {
	return "=" + name
}

func insideTmux() bool {
	return os.Getenv("TMUX") != ""
}
