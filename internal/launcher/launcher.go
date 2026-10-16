package launcher

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/lukasmetzner/tobu/internal/finder"
	"github.com/lukasmetzner/tobu/internal/multiplexer"
	"github.com/lukasmetzner/tobu/internal/picker"
)

type Launcher struct {
	Finder      finder.Finder
	Multiplexer multiplexer.Multiplexer
	Editor      []string
}

func (l Launcher) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return l.pick(ctx)
	}

	dir, err := l.resolve(ctx, args)
	if err != nil {
		return err
	}
	return l.open(ctx, target{session: l.Multiplexer.SessionName(dir), dir: dir})
}

func (l Launcher) pick(ctx context.Context) error {
	sessions, err := l.Multiplexer.Sessions(ctx)
	if err != nil {
		return err
	}

	current, err := l.Multiplexer.CurrentSession(ctx)
	if err != nil {
		return err
	}

	dirs, err := l.Finder.Directories(ctx)
	if err != nil {
		return err
	}

	targets := collect(sessions, current, dirs, l.Multiplexer.SessionName)
	if len(targets) == 0 {
		return errors.New("no open sessions and no directories to pick from")
	}

	items := make([]picker.Item, len(targets))
	for i, t := range targets {
		items[i] = t.item()
	}

	selected, err := picker.Run(ctx, items)
	if err != nil {
		return err
	}
	return l.open(ctx, targets[selected])
}

func (l Launcher) resolve(ctx context.Context, args []string) (string, error) {
	if len(args) == 1 && isDir(args[0]) {
		return filepath.Abs(args[0])
	}
	return l.Finder.Query(ctx, args)
}

func (l Launcher) open(ctx context.Context, t target) error {
	if t.dir != "" {
		if err := l.Finder.Visit(ctx, t.dir); err != nil {
			return err
		}
	}
	if t.running {
		return l.Multiplexer.Attach(t.session)
	}
	return l.Multiplexer.Open(multiplexer.Session{
		Name:    t.session,
		Dir:     t.dir,
		Command: l.Editor,
	})
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
