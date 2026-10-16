package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func Output(ctx context.Context, binary string, args ...string) ([]byte, error) {
	out, err := exec.CommandContext(ctx, binary, args...).Output()
	if err == nil {
		return out, nil
	}

	name := strings.Join(append([]string{filepath.Base(binary)}, args[:min(len(args), 1)]...), " ")
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
		message := strings.TrimPrefix(string(bytes.TrimSpace(exitErr.Stderr)), filepath.Base(binary)+": ")
		return nil, fmt.Errorf("%s: %s", name, message)
	}
	return nil, fmt.Errorf("%s: %w", name, err)
}

func Lines(out []byte) []string {
	var lines []string
	for line := range strings.Lines(string(out)) {
		if line = strings.TrimRight(line, "\r\n"); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
