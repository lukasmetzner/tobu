package finder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/lukasmetzner/tobu/internal/command"
)

type ZoxideConfig struct {
	Binary string `toml:"binary"`
}

type Zoxide struct {
	binary string
}

func NewZoxide(cfg ZoxideConfig) (*Zoxide, error) {
	path, err := exec.LookPath(cfg.Binary)
	if err != nil {
		return nil, fmt.Errorf("zoxide is required: %w", err)
	}
	return &Zoxide{binary: path}, nil
}

func (z *Zoxide) Directories(ctx context.Context) ([]Directory, error) {
	out, err := command.Output(ctx, z.binary, "query", "--list", "--score")
	if err != nil {
		return nil, err
	}
	return parseEntries(out)
}

func (z *Zoxide) Query(ctx context.Context, keywords []string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	args := append([]string{"query", "--exclude", pwd, "--"}, keywords...)
	out, err := command.Output(ctx, z.binary, args...)
	if err != nil {
		if strings.Contains(err.Error(), "no match found") {
			return "", fmt.Errorf("no zoxide match for %q", strings.Join(keywords, " "))
		}
		return "", err
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}

func (z *Zoxide) Visit(ctx context.Context, dir string) error {
	_, err := command.Output(ctx, z.binary, "add", "--", dir)
	return err
}

func parseEntries(out []byte) ([]Directory, error) {
	var dirs []Directory
	for _, line := range command.Lines(out) {
		rawScore, path, ok := strings.Cut(strings.TrimLeft(line, " "), " ")
		if !ok {
			return nil, fmt.Errorf("unexpected zoxide output %q", line)
		}
		score, err := strconv.ParseFloat(rawScore, 64)
		if err != nil {
			return nil, fmt.Errorf("unexpected zoxide score in %q: %w", line, err)
		}
		dirs = append(dirs, Directory{Path: path, Score: score})
	}
	return dirs, nil
}
