package finder

import (
	"context"
	"fmt"
)

type Kind string

const KindZoxide Kind = "zoxide"

type Config struct {
	Kind   Kind         `toml:"kind"`
	Zoxide ZoxideConfig `toml:"zoxide"`
}

type Directory struct {
	Path  string
	Score float64
}

type Finder interface {
	Directories(ctx context.Context) ([]Directory, error)
	Query(ctx context.Context, keywords []string) (string, error)
	Visit(ctx context.Context, dir string) error
}

func New(cfg Config) (Finder, error) {
	switch cfg.Kind {
	case KindZoxide:
		return NewZoxide(cfg.Zoxide)
	default:
		return nil, fmt.Errorf("unsupported finder kind %q", cfg.Kind)
	}
}
