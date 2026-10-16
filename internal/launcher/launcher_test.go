package launcher

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/lukasmetzner/tobu/internal/finder"
	"github.com/lukasmetzner/tobu/internal/multiplexer"
)

type fakeFinder struct {
	matches map[string]string
	queries [][]string
	visited []string
}

func (f *fakeFinder) Directories(context.Context) ([]finder.Directory, error) {
	return nil, nil
}

func (f *fakeFinder) Query(_ context.Context, keywords []string) (string, error) {
	f.queries = append(f.queries, keywords)
	if dir, ok := f.matches[keywords[0]]; ok {
		return dir, nil
	}
	return "", errors.New("no match")
}

func (f *fakeFinder) Visit(_ context.Context, dir string) error {
	f.visited = append(f.visited, dir)
	return nil
}

type fakeMultiplexer struct {
	opened []multiplexer.Session
}

func (m *fakeMultiplexer) SessionName(dir string) string {
	return filepath.Base(dir)
}

func (m *fakeMultiplexer) Sessions(context.Context) ([]string, error) {
	return nil, nil
}

func (m *fakeMultiplexer) CurrentSession(context.Context) (string, error) {
	return "", nil
}

func (m *fakeMultiplexer) Open(s multiplexer.Session) error {
	m.opened = append(m.opened, s)
	return nil
}

func (m *fakeMultiplexer) Attach(string) error {
	return errors.New("unexpected attach")
}

func newTestLauncher(matches map[string]string) (Launcher, *fakeFinder, *fakeMultiplexer) {
	f := &fakeFinder{matches: matches}
	m := &fakeMultiplexer{}
	return Launcher{Finder: f, Multiplexer: m, Editor: []string{"nvim"}}, f, m
}

func TestRunWithKeywordsQueriesFinder(t *testing.T) {
	l, f, m := newTestLauncher(map[string]string{"hccm": "/src/github.com/hccm"})
	if err := l.Run(context.Background(), []string{"hccm", "go"}); err != nil {
		t.Fatal(err)
	}

	if want := [][]string{{"hccm", "go"}}; !reflect.DeepEqual(f.queries, want) {
		t.Errorf("queries = %v, want %v", f.queries, want)
	}
	if want := []string{"/src/github.com/hccm"}; !slices.Equal(f.visited, want) {
		t.Errorf("visited = %v, want %v", f.visited, want)
	}
	want := []multiplexer.Session{{Name: "hccm", Dir: "/src/github.com/hccm", Command: []string{"nvim"}}}
	if !reflect.DeepEqual(m.opened, want) {
		t.Errorf("opened = %+v, want %+v", m.opened, want)
	}
}

func TestRunWithExistingDirectorySkipsQuery(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "hccm"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)

	l, f, m := newTestLauncher(map[string]string{"hccm": "/src/github.com/hccm"})
	if err := l.Run(context.Background(), []string{"hccm"}); err != nil {
		t.Fatal(err)
	}

	if len(f.queries) != 0 {
		t.Errorf("expected no queries, got %v", f.queries)
	}
	want := []multiplexer.Session{{Name: "hccm", Dir: filepath.Join(root, "hccm"), Command: []string{"nvim"}}}
	if !reflect.DeepEqual(m.opened, want) {
		t.Errorf("opened = %+v, want %+v", m.opened, want)
	}
}

func TestRunWithoutMatchFails(t *testing.T) {
	l, f, m := newTestLauncher(nil)
	if err := l.Run(context.Background(), []string{"nothing"}); err == nil {
		t.Fatal("expected error")
	}
	if len(f.visited) != 0 || len(m.opened) != 0 {
		t.Errorf("expected nothing to be opened, visited=%v opened=%v", f.visited, m.opened)
	}
}
