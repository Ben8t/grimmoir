package git

import (
	"fmt"
	"strings"
	"time"
)

type Runner interface {
	Run(dir string, name string, args ...string) error
}

type Syncer struct {
	runner Runner
	now    func() time.Time
}

func NewSyncer(runner Runner) *Syncer {
	return &Syncer{runner: runner, now: time.Now}
}

func (s *Syncer) Sync(repoPath string) error {
	if err := s.ensureRepo(repoPath); err != nil {
		return err
	}

	if err := s.runner.Run(repoPath, "git", "add", "--", "*.md"); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	msg := fmt.Sprintf("Grimmoir sync: %s", s.now().Format(time.RFC3339))
	if err := s.runner.Run(repoPath, "git", "commit", "-m", msg); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	if err := s.runner.Run(repoPath, "git", "pull", "--rebase"); err != nil {
		return fmt.Errorf("git pull --rebase: %w", err)
	}

	if err := s.runner.Run(repoPath, "git", "push"); err != nil {
		return fmt.Errorf("git push: %w", err)
	}

	return nil
}

func (s *Syncer) ensureRepo(repoPath string) error {
	err := s.runner.Run(repoPath, "git", "rev-parse", "--is-inside-work-tree")
	if err == nil {
		return nil
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not a git repository") {
		return fmt.Errorf("git rev-parse: %w", err)
	}
	if err := s.runner.Run(repoPath, "git", "init"); err != nil {
		return fmt.Errorf("git init: %w", err)
	}
	return nil
}
