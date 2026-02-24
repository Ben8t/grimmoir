package git

import (
	"errors"
	"testing"
)

type fakeRunner struct {
	calls []string
	fail  map[string]error
}

func (f *fakeRunner) Run(name string, args ...string) error {
	call := name
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

	if err := s.Sync(); err != nil {
		t.Fatalf("sync: %v", err)
	}

	if len(r.calls) != 4 {
		t.Fatalf("expected 4 git calls, got %d", len(r.calls))
	}
	if r.calls[0] != "git add ." {
		t.Fatalf("unexpected first call: %q", r.calls[0])
	}
	if r.calls[2] != "git pull --rebase" {
		t.Fatalf("unexpected third call: %q", r.calls[2])
	}
	if r.calls[3] != "git push" {
		t.Fatalf("unexpected fourth call: %q", r.calls[3])
	}
}

func TestSyncStopsOnError(t *testing.T) {
	r := &fakeRunner{fail: map[string]error{"git pull --rebase": errors.New("conflict")}}
	s := NewSyncer(r)

	err := s.Sync()
	if err == nil {
		t.Fatal("expected sync error")
	}

	if len(r.calls) != 3 {
		t.Fatalf("expected to stop at failing command, got calls: %#v", r.calls)
	}
}
