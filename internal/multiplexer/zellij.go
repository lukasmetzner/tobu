package multiplexer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/lukasmetzner/tobu/internal/command"
)

var kdlEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`)

type Zellij struct {
	binary    string
	nameDepth int
}

type zellijSession struct {
	name   string
	exited bool
}

func (z *Zellij) SessionName(dir string) string {
	return sessionName(dir, z.nameDepth)
}

func (z *Zellij) Sessions(ctx context.Context) ([]string, error) {
	all, err := z.allSessions(ctx)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, s := range all {
		if !s.exited {
			names = append(names, s.name)
		}
	}
	return names, nil
}

func (z *Zellij) CurrentSession(context.Context) (string, error) {
	if !insideZellij() {
		return "", nil
	}
	return os.Getenv("ZELLIJ_SESSION_NAME"), nil
}

func (z *Zellij) Open(ctx context.Context, s Session) error {
	all, err := z.allSessions(ctx)
	if err != nil {
		return err
	}
	exists := slices.ContainsFunc(all, func(existing zellijSession) bool { return existing.name == s.Name })
	if err := z.createBackground(s, !exists); err != nil {
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

// createBackground starts a new session from the layout, or resurrects an exited one with its own layout.
// Zellij adds the layout to the current session instead when it sees the variables of a running client.
func (z *Zellij) createBackground(s Session, withLayout bool) error {
	cmd := exec.Command(z.binary, zellijCreateArgs(s, withLayout)...)
	cmd.Dir = s.Dir
	cmd.Env = slices.DeleteFunc(os.Environ(), func(kv string) bool { return strings.HasPrefix(kv, "ZELLIJ") })
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s attach: %w", filepath.Base(z.binary), err)
	}
	return nil
}

func (z *Zellij) allSessions(ctx context.Context) ([]zellijSession, error) {
	out, err := command.Output(ctx, z.binary, "list-sessions", "--no-formatting")
	if err != nil {
		if strings.Contains(err.Error(), "No active zellij sessions") {
			return nil, nil
		}
		return nil, err
	}
	return parseZellijSessions(out), nil
}

func parseZellijSessions(out []byte) []zellijSession {
	var sessions []zellijSession
	for _, line := range command.Lines(out) {
		name, status, ok := strings.Cut(line, " [Created ")
		if ok {
			sessions = append(sessions, zellijSession{name: name, exited: strings.Contains(status, "(EXITED")})
		}
	}
	return sessions
}

func zellijCreateArgs(s Session, withLayout bool) []string {
	var args []string
	if withLayout {
		args = append(args, "--layout-string", zellijLayout(s))
	}
	return append(args, "attach", "--create-background", "--force-run-commands", s.Name)
}

func zellijLayout(s Session) string {
	var b strings.Builder
	fmt.Fprintf(&b, "layout {\n    cwd %s\n", kdlString(s.Dir))
	b.WriteString(`    default_tab_template {
        pane size=1 borderless=true {
            plugin location="zellij:tab-bar"
        }
        children
        pane size=1 borderless=true {
            plugin location="zellij:status-bar"
        }
    }
`)
	for i, c := range s.Commands {
		focus := ""
		if i == 0 {
			focus = " focus=true"
		}
		fmt.Fprintf(&b, "    tab name=%s%s {\n        pane command=%s", kdlString(filepath.Base(c[0])), focus, kdlString(c[0]))
		if len(c) > 1 {
			b.WriteString(" {\n            args")
			for _, arg := range c[1:] {
				b.WriteString(" " + kdlString(arg))
			}
			b.WriteString("\n        }")
		}
		b.WriteString("\n    }\n")
	}
	b.WriteString("}\n")
	return b.String()
}

func kdlString(s string) string {
	return `"` + kdlEscaper.Replace(s) + `"`
}

func insideZellij() bool {
	return os.Getenv("ZELLIJ") != ""
}
