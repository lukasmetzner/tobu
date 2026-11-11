package multiplexer

import (
	"slices"
	"testing"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name   string
		cfg    Config
		kind   Kind
		binary string
		err    bool
	}{
		{name: "defaults to zellij", cfg: Config{}, kind: KindZellij, binary: "zellij"},
		{name: "infers tmux from binary", cfg: Config{Binary: "/usr/local/bin/tmux"}, kind: KindTmux, binary: "/usr/local/bin/tmux"},
		{name: "binary defaults to kind", cfg: Config{Kind: KindTmux}, kind: KindTmux, binary: "tmux"},
		{name: "explicit kind with custom binary", cfg: Config{Kind: KindZellij, Binary: "/opt/zj"}, kind: KindZellij, binary: "/opt/zj"},
		{name: "unknown binary without kind", cfg: Config{Binary: "/opt/zj"}, err: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kind, binary, err := resolve(tt.cfg)
			if tt.err {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if kind != tt.kind || binary != tt.binary {
				t.Errorf("got (%q, %q), want (%q, %q)", kind, binary, tt.kind, tt.binary)
			}
		})
	}
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	if _, err := New(Config{Kind: "screen", Binary: "/bin/sh"}, SessionConfig{NameDepth: 1}); err == nil {
		t.Error("expected error for unsupported kind")
	}
	if _, err := New(Config{Kind: KindTmux, Binary: "/bin/sh"}, SessionConfig{NameDepth: 0}); err == nil {
		t.Error("expected error for name depth below 1")
	}
}

func TestSessionName(t *testing.T) {
	tests := []struct {
		dir    string
		depth  int
		zellij string
		tmux   string
	}{
		{dir: "/home/user/projects/github.com/tobu", depth: 1, zellij: "tobu", tmux: "tobu"},
		{dir: "/home/user/projects/github.com/tobu", depth: 2, zellij: "github.com_tobu", tmux: "github_com/tobu"},
		{dir: "/home/user/projects/github.com/tobu/", depth: 3, zellij: "projects_github.com_tobu", tmux: "projects/github_com/tobu"},
		{dir: "/tmp", depth: 5, zellij: "tmp", tmux: "tmp"},
		{dir: "/home/user/my project:v2", depth: 1, zellij: "my project:v2", tmux: "my project_v2"},
	}
	for _, tt := range tests {
		if got := (&Zellij{nameDepth: tt.depth}).SessionName(tt.dir); got != tt.zellij {
			t.Errorf("zellij SessionName(%q, %d) = %q, want %q", tt.dir, tt.depth, got, tt.zellij)
		}
		if got := (&Tmux{nameDepth: tt.depth}).SessionName(tt.dir); got != tt.tmux {
			t.Errorf("tmux SessionName(%q, %d) = %q, want %q", tt.dir, tt.depth, got, tt.tmux)
		}
	}
}

func TestParseZellijSessions(t *testing.T) {
	out := []byte("tobu [Created 2h 3m ago] (current)\n" +
		"github.com_hccm [Created 1h ago] \n" +
		"old-session [Created 3months ago] (EXITED - attach to resurrect)\n" +
		"has space [Created 5s ago] \n")
	got := parseZellijSessions(out)
	want := []zellijSession{
		{name: "tobu"},
		{name: "github.com_hccm"},
		{name: "old-session", exited: true},
		{name: "has space"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestZellijCreateArgs(t *testing.T) {
	s := Session{Name: "tobu", Dir: "/src/tobu", Commands: [][]string{{"nvim", "."}, {"lazygit"}}}

	if got, want := zellijCreateArgs(s, false), []string{"attach", "--create-background", "--force-run-commands", "tobu"}; !slices.Equal(got, want) {
		t.Errorf("args = %v, want %v", got, want)
	}

	got := zellijCreateArgs(s, true)
	if want := []string{"--layout-string", zellijLayout(s), "attach", "--create-background", "--force-run-commands", "tobu"}; !slices.Equal(got, want) {
		t.Errorf("args = %v, want %v", got, want)
	}
}

func TestZellijLayout(t *testing.T) {
	s := Session{Name: "tobu", Dir: `/src/"my" tobu`, Commands: [][]string{{"/usr/bin/nvim", "."}, {"lazygit"}}}
	want := `layout {
    cwd "/src/\"my\" tobu"
    default_tab_template {
        pane size=1 borderless=true {
            plugin location="zellij:tab-bar"
        }
        children
        pane size=1 borderless=true {
            plugin location="zellij:status-bar"
        }
    }
    tab name="nvim" focus=true {
        pane command="/usr/bin/nvim" {
            args "."
        }
    }
    tab name="lazygit" {
        pane command="lazygit"
    }
}
`
	if got := zellijLayout(s); got != want {
		t.Errorf("layout:\n%s\nwant:\n%s", got, want)
	}
}

func TestTmuxArgs(t *testing.T) {
	s := Session{Name: "tobu", Dir: "/src/tobu", Commands: [][]string{{"nvim", "."}, {"lazygit"}}}

	if got, want := tmuxNewSessionArgs(s), []string{"new-session", "-d", "-s", "tobu", "-c", "/src/tobu", "--", "nvim", "."}; !slices.Equal(got, want) {
		t.Errorf("new-session args = %v, want %v", got, want)
	}
	if got, want := tmuxNewWindowArgs(s, s.Commands[1]), []string{"new-window", "-d", "-t", "=tobu:", "-c", "/src/tobu", "--", "lazygit"}; !slices.Equal(got, want) {
		t.Errorf("new-window args = %v, want %v", got, want)
	}
}
