package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"grimmoir/internal/model"
	"grimmoir/internal/service"
)

type fakeRepo struct {
	skills      []model.Skill
	deletedName string
}

func (f *fakeRepo) ListSkills() ([]model.Skill, error) {
	return f.skills, nil
}

func (f *fakeRepo) SaveSkill(skill model.Skill) (string, error) {
	return "", nil
}

func (f *fakeRepo) DeleteSkill(name string) (string, error) {
	f.deletedName = name
	out := make([]model.Skill, 0, len(f.skills))
	for _, s := range f.skills {
		if s.Name == name {
			continue
		}
		out = append(out, s)
	}
	f.skills = out
	return "deleted.md", nil
}

func TestDeleteKeyRemovesSelectedSkill(t *testing.T) {
	repo := &fakeRepo{skills: []model.Skill{{Name: "One", Body: "one"}, {Name: "Two", Body: "two"}}}
	svc := service.New(repo, nil)

	m := modelUI{
		svc:      svc,
		all:      repo.skills,
		filtered: repo.skills,
		selected: 1,
		stack:    []string{"Two"},
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	got := updated.(modelUI)

	if repo.deletedName != "Two" {
		t.Fatalf("expected deleted name Two, got %q", repo.deletedName)
	}
	if len(got.filtered) != 1 || got.filtered[0].Name != "One" {
		t.Fatalf("unexpected filtered list after delete: %#v", got.filtered)
	}
	if len(got.stack) != 0 {
		t.Fatalf("expected removed item from stack, got %#v", got.stack)
	}
}
