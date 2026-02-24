package git

import (
	"fmt"
	"time"
)

type Runner interface {
	Run(name string, args ...string) error
}

type Syncer struct {
	runner Runner
	now    func() time.Time
}

func NewSyncer(runner Runner) *Syncer {
	return &Syncer{runner: runner, now: time.Now}
}

func (s *Syncer) Sync() error {
	if err := s.runner.Run("git", "add", "."); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	msg := fmt.Sprintf("Grimmoir sync: %s", s.now().Format(time.RFC3339))
	if err := s.runner.Run("git", "commit", "-m", msg); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	if err := s.runner.Run("git", "pull", "--rebase"); err != nil {
		return fmt.Errorf("git pull --rebase: %w", err)
	}

	if err := s.runner.Run("git", "push"); err != nil {
		return fmt.Errorf("git push: %w", err)
	}

	return nil
}
