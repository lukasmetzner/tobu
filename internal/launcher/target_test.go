package launcher

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/lukasmetzner/tobu/internal/finder"
	"github.com/lukasmetzner/tobu/internal/picker"
)

func TestCollect(t *testing.T) {
	dirs := []finder.Directory{
		{Path: "/src/github.com/tobu", Score: 50},
		{Path: "/src/github.com/hccm", Score: 40},
		{Path: "/src/gitlab.com/tobu", Score: 30},
		{Path: "/src/github.com/csi", Score: 20},
		{Path: "/src/github.com/current", Score: 10},
	}
	sessions := []string{"zeta", "csi", "current", "tobu", "adhoc"}

	got := collect(sessions, "current", dirs, filepath.Base)
	want := []target{
		{session: "tobu", dir: "/src/github.com/tobu", score: 50, running: true},
		{session: "csi", dir: "/src/github.com/csi", score: 20, running: true},
		{session: "adhoc", running: true},
		{session: "zeta", running: true},
		{session: "hccm", dir: "/src/github.com/hccm", score: 40},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("collect() =\n%+v\nwant\n%+v", got, want)
	}
}

func TestCollectWithoutSessions(t *testing.T) {
	dirs := []finder.Directory{
		{Path: "/src/a", Score: 2},
		{Path: "/src/b", Score: 1},
	}
	got := collect(nil, "", dirs, filepath.Base)
	want := []target{
		{session: "a", dir: "/src/a", score: 2},
		{session: "b", dir: "/src/b", score: 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("collect() = %+v, want %+v", got, want)
	}
}

func TestTargetItem(t *testing.T) {
	session := target{session: "tobu", dir: "/src/tobu", running: true}.item()
	if want := (picker.Item{Label: "tobu", Tag: sessionTag, Accent: true}); session != want {
		t.Errorf("session item = %+v, want %+v", session, want)
	}

	dir := target{session: "tobu", dir: "/src/tobu", score: 12.25}.item()
	if want := (picker.Item{Label: "/src/tobu", Tag: "12.2"}); dir != want {
		t.Errorf("directory item = %+v, want %+v", dir, want)
	}
}
