package multiplexer

import (
	"slices"
	"testing"
)

func TestNewRejectsInvalidConfig(t *testing.T) {
	if _, err := New(Config{Binary: "/nonexistent/zellij"}, SessionConfig{NameDepth: 1}); err == nil {
		t.Error("expected error for missing binary")
	}
	if _, err := New(Config{Binary: "/bin/sh"}, SessionConfig{NameDepth: 0}); err == nil {
		t.Error("expected error for name depth below 1")
	}
}

func TestSessionName(t *testing.T) {
	tests := []struct {
		dir   string
		depth int
		want  string
	}{
		{dir: "/home/user/projects/github.com/tobu", depth: 1, want: "tobu"},
		{dir: "/home/user/projects/github.com/tobu", depth: 2, want: "github.com_tobu"},
		{dir: "/home/user/projects/github.com/tobu/", depth: 3, want: "projects_github.com_tobu"},
		{dir: "/tmp", depth: 5, want: "tmp"},
		{dir: "/home/user/my project:v2", depth: 1, want: "my project:v2"},
	}
	for _, tt := range tests {
		if got := (&Zellij{nameDepth: tt.depth}).SessionName(tt.dir); got != tt.want {
			t.Errorf("SessionName(%q, %d) = %q, want %q", tt.dir, tt.depth, got, tt.want)
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
