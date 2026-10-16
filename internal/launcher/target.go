package launcher

import (
	"slices"
	"strconv"

	"github.com/lukasmetzner/tobu/internal/finder"
	"github.com/lukasmetzner/tobu/internal/picker"
)

const sessionTag = "session"

type target struct {
	session string
	dir     string
	score   float64
	running bool
}

func collect(sessions []string, current string, dirs []finder.Directory, sessionName func(dir string) string) []target {
	running := make(map[string]bool, len(sessions))
	for _, name := range sessions {
		running[name] = true
	}

	var sessionTargets, dirTargets []target
	listed := map[string]bool{current: true}
	for _, dir := range dirs {
		name := sessionName(dir.Path)
		switch {
		case !running[name]:
			dirTargets = append(dirTargets, target{session: name, dir: dir.Path, score: dir.Score})
		case !listed[name]:
			listed[name] = true
			sessionTargets = append(sessionTargets, target{session: name, dir: dir.Path, score: dir.Score, running: true})
		}
	}

	for _, name := range slices.Sorted(slices.Values(sessions)) {
		if !listed[name] {
			sessionTargets = append(sessionTargets, target{session: name, running: true})
		}
	}
	return append(sessionTargets, dirTargets...)
}

func (t target) item() picker.Item {
	if t.running {
		return picker.Item{Label: t.session, Tag: sessionTag, Accent: true}
	}
	return picker.Item{Label: t.dir, Tag: strconv.FormatFloat(t.score, 'f', 1, 64)}
}
