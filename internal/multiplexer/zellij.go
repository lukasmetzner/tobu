package multiplexer

import (
	"context"
	"os"
	"strings"

	"github.com/lukasmetzner/tobu/internal/command"
)

type Zellij struct {
	binary    string
	nameDepth int
}

func (z *Zellij) SessionName(dir string) string {
	return strings.ReplaceAll(sessionName(dir, z.nameDepth), "/", "_")
}

func (z *Zellij) Sessions(ctx context.Context) ([]string, error) {
	out, err := command.Output(ctx, z.binary, "list-sessions", "--no-formatting")
	if err != nil {
		if strings.Contains(err.Error(), "No active zellij sessions") {
			return nil, nil
		}
		return nil, err
	}
	return parseZellijSessions(out), nil
}

func (z *Zellij) CurrentSession(context.Context) (string, error) {
	if !insideZellij() {
		return "", nil
	}
	return os.Getenv("ZELLIJ_SESSION_NAME"), nil
}

func (z *Zellij) Open(s Session) error {
	if !insideZellij() {
		return replaceProcess(s.Dir, z.binary, zellijAttachArgs("--create", s)...)
	}
	if err := run(s.Dir, z.binary, zellijAttachArgs("--create-background", s)...); err != nil {
		return err
	}
	return z.Attach(s.Name)
}

func (z *Zellij) Attach(name string) error {
	if !insideZellij() {
		return replaceProcess("", z.binary, "attach", name)
	}
	return run("", z.binary, "action", "switch-session", name)
}

func parseZellijSessions(out []byte) []string {
	var names []string
	for _, line := range command.Lines(out) {
		name, status, ok := strings.Cut(line, " [Created ")
		if ok && !strings.Contains(status, "(EXITED") {
			names = append(names, name)
		}
	}
	return names
}

func zellijAttachArgs(mode string, s Session) []string {
	return append([]string{"attach", mode, "--force-run-commands", s.Name, "--"}, s.Command...)
}

func insideZellij() bool {
	return os.Getenv("ZELLIJ") != ""
}
