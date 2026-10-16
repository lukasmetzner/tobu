package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/lukasmetzner/tobu/internal/config"
	"github.com/lukasmetzner/tobu/internal/finder"
	"github.com/lukasmetzner/tobu/internal/launcher"
	"github.com/lukasmetzner/tobu/internal/multiplexer"
	"github.com/lukasmetzner/tobu/internal/picker"
)

func main() {
	err := run(context.Background())
	switch {
	case err == nil:
	case errors.Is(err, picker.ErrAborted):
		os.Exit(130)
	default:
		fmt.Fprintf(os.Stderr, "tobu: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if _, err := exec.LookPath(cfg.Editor.Binary); err != nil {
		return fmt.Errorf("editor: %w", err)
	}

	f, err := finder.New(cfg.Finder)
	if err != nil {
		return err
	}

	mux, err := multiplexer.New(cfg.Multiplexer, cfg.Session)
	if err != nil {
		return err
	}

	return launcher.Launcher{
		Finder:      f,
		Multiplexer: mux,
		Editor:      cfg.Editor.Command(),
	}.Run(ctx, os.Args[1:])
}
