package config

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/lukasmetzner/tobu/internal/finder"
	"github.com/lukasmetzner/tobu/internal/multiplexer"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadWritesDefaultFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tobu", "config.toml")
	t.Setenv(envPath, path)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg, Default()) {
		t.Errorf("got %+v, want defaults %+v", cfg, Default())
	}

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(written, defaultFile) {
		t.Errorf("written config differs from embedded default:\n%s", written)
	}
}

func TestLoadKeepsExistingFile(t *testing.T) {
	content := "[session]\nname_depth = 2\n"
	path := writeConfig(t, content)
	t.Setenv(envPath, path)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Session.NameDepth != 2 {
		t.Errorf("name_depth = %d, want 2", cfg.Session.NameDepth)
	}

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(written) != content {
		t.Errorf("existing config was modified:\n%s", written)
	}
}

func TestLoadFileMissing(t *testing.T) {
	if _, err := LoadFile(filepath.Join(t.TempDir(), "missing.toml")); err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestLoadFileOverridesDefaults(t *testing.T) {
	path := writeConfig(t, `
[multiplexer]
binary = "/usr/bin/tmux"

[session]
name_depth = 2

[editor]
binary = "emacs"
args = ["-nw"]
`)
	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	want := Config{
		Finder: finder.Config{
			Kind:   finder.KindZoxide,
			Zoxide: finder.ZoxideConfig{Binary: "zoxide"},
		},
		Multiplexer: multiplexer.Config{Binary: "/usr/bin/tmux"},
		Session:     multiplexer.SessionConfig{NameDepth: 2},
		Editor:      Editor{Binary: "emacs", Args: []string{"-nw"}},
	}
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("got %+v, want %+v", cfg, want)
	}
	if got := cfg.Editor.Command(); !reflect.DeepEqual(got, []string{"emacs", "-nw"}) {
		t.Errorf("editor command = %v", got)
	}
}

func TestLoadFileRejectsUnknownKeys(t *testing.T) {
	path := writeConfig(t, `
[editor]
bin = "vim"
`)
	_, err := LoadFile(path)
	if err == nil || !strings.Contains(err.Error(), "editor.bin") {
		t.Fatalf("expected unknown key error, got %v", err)
	}
}

func TestPathFromEnvironment(t *testing.T) {
	t.Setenv(envPath, "/tmp/tobu.toml")
	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if path != "/tmp/tobu.toml" {
		t.Errorf("path = %q", path)
	}
}
