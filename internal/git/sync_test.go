package git

import (
	"errors"
	"testing"
)

type fakeRunner struct {
	calls []string
	fail  map[string]error
}

func (f *fakeRunner) Run(dir string, name string, args ...string) error {
	call := dir + " :: " + name
	for _, a := range args {
		call += " " + a
	}
	f.calls = append(f.calls, call)
	if err := f.fail[call]; err != nil {
		return err
	}
	return nil
}

func TestSyncRunsExpectedGitCommands(t *testing.T) {
	r := &fakeRunner{fail: map[string]error{}}
	s := NewSyncer(r)

	if err := s.Sync("/repo/prompts"); err != nil {
		t.Fatalf("sync: %v", err)
	}

	if len(r.calls) != 5 {
		t.Fatalf("expected 5 git calls, got %d", len(r.calls))
	}
	if r.calls[0] != "/repo/prompts :: git rev-parse --is-inside-work-tree" {
		t.Fatalf("unexpected precheck call: %q", r.calls[0])
	}
	if r.calls[1] != "/repo/prompts :: git add -- *.md" {
		t.Fatalf("unexpected add call: %q", r.calls[1])
	}
	if r.calls[3] != "/repo/prompts :: git pull --rebase" {
		t.Fatalf("unexpected pull call: %q", r.calls[3])
	}
	if r.calls[4] != "/repo/prompts :: git push" {
		t.Fatalf("unexpected push call: %q", r.calls[4])
	}
}

func TestSyncStopsOnError(t *testing.T) {
	r := &fakeRunner{fail: map[string]error{"/repo/prompts :: git pull --rebase": errors.New("conflict")}}
	s := NewSyncer(r)

	err := s.Sync("/repo/prompts")
	if err == nil {
		t.Fatal("expected sync error")
	}

	if len(r.calls) != 4 {
		t.Fatalf("expected to stop at failing command, got calls: %#v", r.calls)
	}
}

func TestSyncInitializesRepoWhenMissing(t *testing.T) {
	r := &fakeRunner{fail: map[string]error{"/repo/prompts :: git rev-parse --is-inside-work-tree": errors.New("not a git repository")}}
	s := NewSyncer(r)

	if err := s.Sync("/repo/prompts"); err != nil {
		t.Fatalf("sync: %v", err)
	}

	if len(r.calls) < 2 {
		t.Fatalf("expected at least 2 calls, got %d", len(r.calls))
	}
	if r.calls[0] != "/repo/prompts :: git rev-parse --is-inside-work-tree" {
		t.Fatalf("unexpected first call: %q", r.calls[0])
	}
	if r.calls[1] != "/repo/prompts :: git init" {
		t.Fatalf("unexpected second call: %q", r.calls[1])
	}
}
