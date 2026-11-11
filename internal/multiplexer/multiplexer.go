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

type Config struct {
	Binary string `toml:"binary"`
}

type SessionConfig struct {
	NameDepth int `toml:"name_depth"`
}

type Session struct {
	Name     string
	Dir      string
	Commands [][]string
}

type Multiplexer interface {
	SessionName(dir string) string
	Sessions(ctx context.Context) ([]string, error)
	CurrentSession(ctx context.Context) (string, error)
	Open(ctx context.Context, s Session) error
	Attach(name string) error
}

func New(cfg Config, session SessionConfig) (Multiplexer, error) {
	if session.NameDepth < 1 {
		return nil, fmt.Errorf("session.name_depth must be at least 1, got %d", session.NameDepth)
	}

	path, err := exec.LookPath(cfg.Binary)
	if err != nil {
		return nil, fmt.Errorf("zellij: %w", err)
	}
	return &Zellij{binary: path, nameDepth: session.NameDepth}, nil
}

func sessionName(dir string, depth int) string {
	parts := strings.Split(strings.Trim(filepath.Clean(dir), "/"), "/")
	return strings.Join(parts[max(len(parts)-depth, 0):], "_")
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
