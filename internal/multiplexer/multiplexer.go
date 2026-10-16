package multiplexer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
)

type Kind string

const (
	KindZellij Kind = "zellij"
	KindTmux   Kind = "tmux"
)

type Config struct {
	Kind   Kind   `toml:"kind"`
	Binary string `toml:"binary"`
}

type SessionConfig struct {
	NameDepth int `toml:"name_depth"`
}

type Session struct {
	Name    string
	Dir     string
	Command []string
}

type Multiplexer interface {
	SessionName(dir string) string
	Sessions(ctx context.Context) ([]string, error)
	CurrentSession(ctx context.Context) (string, error)
	Open(s Session) error
	Attach(name string) error
}

func New(cfg Config, session SessionConfig) (Multiplexer, error) {
	if session.NameDepth < 1 {
		return nil, fmt.Errorf("session.name_depth must be at least 1, got %d", session.NameDepth)
	}

	kind, binary, err := resolve(cfg)
	if err != nil {
		return nil, err
	}

	path, err := exec.LookPath(binary)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", kind, err)
	}

	switch kind {
	case KindZellij:
		return &Zellij{binary: path, nameDepth: session.NameDepth}, nil
	case KindTmux:
		return &Tmux{binary: path, nameDepth: session.NameDepth}, nil
	default:
		return nil, fmt.Errorf("unsupported multiplexer kind %q", kind)
	}
}

func resolve(cfg Config) (Kind, string, error) {
	switch {
	case cfg.Kind == "" && cfg.Binary == "":
		return KindZellij, string(KindZellij), nil
	case cfg.Kind == "":
		kind := Kind(filepath.Base(cfg.Binary))
		if kind != KindZellij && kind != KindTmux {
			return "", "", fmt.Errorf("cannot infer multiplexer kind from binary %q, set multiplexer.kind", cfg.Binary)
		}
		return kind, cfg.Binary, nil
	case cfg.Binary == "":
		return cfg.Kind, string(cfg.Kind), nil
	default:
		return cfg.Kind, cfg.Binary, nil
	}
}

func sessionName(dir string, depth int) string {
	parts := strings.Split(strings.Trim(filepath.Clean(dir), "/"), "/")
	return strings.Join(parts[max(len(parts)-depth, 0):], "/")
}

func run(dir, binary string, args ...string) error {
	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", filepath.Base(binary), args[0], err)
	}
	return nil
}

func replaceProcess(dir, binary string, args ...string) error {
	env := os.Environ()
	if dir != "" {
		if err := os.Chdir(dir); err != nil {
			return err
		}
		env = slices.DeleteFunc(env, func(kv string) bool { return strings.HasPrefix(kv, "PWD=") })
		env = append(env, "PWD="+dir)
	}
	argv := append([]string{filepath.Base(binary)}, args...)
	return syscall.Exec(binary, argv, env)
}
