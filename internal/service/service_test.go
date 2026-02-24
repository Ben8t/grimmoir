package service

import (
	"errors"
	"testing"

	"grimmoir/internal/model"
)

type fakeRepo struct {
	listSkills []model.Skill
	listErr    error
	saved      *model.Skill
	saveErr    error
	deletedName string
	deleteErr   error
}

func (f *fakeRepo) ListSkills() ([]model.Skill, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listSkills, nil
}

func (f *fakeRepo) SaveSkill(skill model.Skill) (string, error) {
	if f.saveErr != nil {
		return "", f.saveErr
	}
	f.saved = &skill
	return "saved.md", nil
}

func (f *fakeRepo) DeleteSkill(name string) (string, error) {
	if f.deleteErr != nil {
		return "", f.deleteErr
	}
	f.deletedName = name
	return "deleted.md", nil
}

func TestComposeIncludesHeadersAndBodiesInOrder(t *testing.T) {
	repo := &fakeRepo{listSkills: []model.Skill{
		{Name: "A", Body: "alpha"},
		{Name: "B", Body: "bravo"},
	}}

	svc := New(repo, nil)

	got, err := svc.Compose([]string{"B", "A"})
	if err != nil {
		t.Fatalf("compose: %v", err)
	}

	want := "# B\n\nbravo\n\n# A\n\nalpha"
	if got != want {
		t.Fatalf("unexpected compose output:\nwant:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestAddFromTextValidatesName(t *testing.T) {
	repo := &fakeRepo{}
	svc := New(repo, nil)

	_, err := svc.AddFromText(model.Skill{Body: "body"})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestAddFromTextPersistsSkill(t *testing.T) {
	repo := &fakeRepo{}
	svc := New(repo, nil)

	_, err := svc.AddFromText(model.Skill{
		Name:        "Prompt Expert",
		Description: "Desc",
		Body:        "content",
	})
	if err != nil {
		t.Fatalf("add from text: %v", err)
	}

	if repo.saved == nil {
		t.Fatal("expected skill to be saved")
	}
	if repo.saved.Version == "" {
		t.Fatal("expected default version to be set")
	}
}

func TestListBubblesRepoError(t *testing.T) {
	repo := &fakeRepo{listErr: errors.New("boom")}
	svc := New(repo, nil)

	_, err := svc.List()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteByNameValidatesInput(t *testing.T) {
	svc := New(&fakeRepo{}, nil)

	_, err := svc.DeleteByName(" ")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteByNameDelegatesToRepo(t *testing.T) {
	repo := &fakeRepo{}
	svc := New(repo, nil)

	path, err := svc.DeleteByName("Prompt Expert")
	if err != nil {
		t.Fatalf("delete by name: %v", err)
	}
	if path != "deleted.md" {
		t.Fatalf("unexpected path: %q", path)
	}
	if repo.deletedName != "Prompt Expert" {
		t.Fatalf("unexpected deleted name: %q", repo.deletedName)
	}
}
