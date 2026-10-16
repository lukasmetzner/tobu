package config

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/lukasmetzner/tobu/internal/finder"
	"github.com/lukasmetzner/tobu/internal/multiplexer"
)

const envPath = "TOBU_CONFIG"

//go:embed default.toml
var defaultFile []byte

type Config struct {
	Finder      finder.Config             `toml:"finder"`
	Multiplexer multiplexer.Config        `toml:"multiplexer"`
	Session     multiplexer.SessionConfig `toml:"session"`
	Editor      Editor                    `toml:"editor"`
}

type Editor struct {
	Binary string   `toml:"binary"`
	Args   []string `toml:"args"`
}

func (e Editor) Command() []string {
	return append([]string{e.Binary}, e.Args...)
}

func Default() Config {
	return Config{
		Finder: finder.Config{
			Kind:   finder.KindZoxide,
			Zoxide: finder.ZoxideConfig{Binary: "zoxide"},
		},
		Session: multiplexer.SessionConfig{NameDepth: 1},
		Editor:  Editor{Binary: "nvim", Args: []string{}},
	}
}

func Path() (string, error) {
	if path := os.Getenv(envPath); path != "" {
		return path, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "tobu", "config.toml"), nil
}

func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}
	if err := writeDefaultFile(path); err != nil {
		return Config{}, fmt.Errorf("write default config %s: %w", path, err)
	}
	return LoadFile(path)
}

func LoadFile(path string) (Config, error) {
	cfg := Default()
	meta, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("config %s: %w", path, err)
	}
	if undecoded := meta.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, len(undecoded))
		for i, key := range undecoded {
			keys[i] = key.String()
		}
		return Config{}, fmt.Errorf("config %s: unknown keys: %s", path, strings.Join(keys, ", "))
	}
	return cfg, nil
}

func writeDefaultFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err := file.Write(defaultFile); err != nil {
		return errors.Join(err, file.Close())
	}
	return file.Close()
}
