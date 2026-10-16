package finder

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseEntries(t *testing.T) {
	out := []byte("2140.0 /home/user/projects\n 544.0 /home/user/my dir\n   0.2 /tmp\n\n")
	dirs, err := parseEntries(out)
	if err != nil {
		t.Fatal(err)
	}
	want := []Directory{
		{Path: "/home/user/projects", Score: 2140},
		{Path: "/home/user/my dir", Score: 544},
		{Path: "/tmp", Score: 0.2},
	}
	if !reflect.DeepEqual(dirs, want) {
		t.Errorf("got %+v, want %+v", dirs, want)
	}
}

func TestParseEntriesEmpty(t *testing.T) {
	dirs, err := parseEntries(nil)
	if err != nil || len(dirs) != 0 {
		t.Fatalf("got %v, %v; want no directories", dirs, err)
	}
}

func TestParseEntriesInvalid(t *testing.T) {
	for _, out := range []string{"/no/score\n", "abc /path\n"} {
		if _, err := parseEntries([]byte(out)); err == nil {
			t.Errorf("expected error for %q", out)
		}
	}
}

func TestNewRejectsUnknownKind(t *testing.T) {
	if _, err := New(Config{Kind: "explorer"}); err == nil {
		t.Fatal("expected error for unsupported finder kind")
	}
}

func fakeZoxide(t *testing.T) (*Zoxide, string) {
	t.Helper()
	dir := t.TempDir()
	argsFile := filepath.Join(dir, "args")
	script := filepath.Join(dir, "zoxide")
	content := "#!/bin/sh\n" +
		"printf '%s\\n' \"$@\" > " + argsFile + "\n" +
		"[ \"$5\" = missing ] && { echo 'zoxide: no match found' >&2; exit 1; }\n" +
		"echo '/home/user/my dir'\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	return &Zoxide{binary: script}, argsFile
}

func TestQuery(t *testing.T) {
	z, argsFile := fakeZoxide(t)
	dir, err := z.Query(context.Background(), []string{"hccm", "go"})
	if err != nil {
		t.Fatal(err)
	}
	if dir != "/home/user/my dir" {
		t.Errorf("dir = %q", dir)
	}

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	args, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	if want := "query\n--exclude\n" + pwd + "\n--\nhccm\ngo\n"; string(args) != want {
		t.Errorf("args = %q, want %q", args, want)
	}
}

func TestQueryNoMatch(t *testing.T) {
	z, _ := fakeZoxide(t)
	_, err := z.Query(context.Background(), []string{"missing"})
	if err == nil || err.Error() != `no zoxide match for "missing"` {
		t.Fatalf("err = %v", err)
	}
}
