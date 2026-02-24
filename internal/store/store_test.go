package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"grimmoir/internal/model"
)

func TestListSkillsParsesMarkdownFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	content := `---
name: "Security Auditor"
description: "Audits Go code for vulnerabilities."
tags: [golang, security]
version: "1.0.0"
---

# Security Auditor Skill
Review code for SQL injection.`

	if err := os.WriteFile(filepath.Join(tmp, "security_auditor.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	s, err := New(tmp)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	skills, err := s.ListSkills()
	if err != nil {
		t.Fatalf("list skills: %v", err)
	}

	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}

	got := skills[0]
	if got.Name != "Security Auditor" {
		t.Fatalf("unexpected name: %q", got.Name)
	}
	if got.Description != "Audits Go code for vulnerabilities." {
		t.Fatalf("unexpected description: %q", got.Description)
	}
	if got.Version != "1.0.0" {
		t.Fatalf("unexpected version: %q", got.Version)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "golang" || got.Tags[1] != "security" {
		t.Fatalf("unexpected tags: %#v", got.Tags)
	}
	if got.Body == "" {
		t.Fatal("expected non-empty body")
	}
}

func TestSaveSkillWritesValidMarkdownFile(t *testing.T) {
	tmp := t.TempDir()
	s, err := New(tmp)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	skill := model.Skill{
		Name:        "Code Reviewer",
		Description: "Reviews pull requests",
		Tags:        []string{"review", "quality"},
		Version:     "0.1.0",
		Body:        "# Code Reviewer\nLook for maintainability and tests.",
	}

	path, err := s.SaveSkill(skill)
	if err != nil {
		t.Fatalf("save skill: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	text := string(b)
	checks := []string{
		"---",
		"name: \"Code Reviewer\"",
		"description: \"Reviews pull requests\"",
		"tags: [review, quality]",
		"version: \"0.1.0\"",
		"# Code Reviewer",
	}
	for _, c := range checks {
		if !strings.Contains(text, c) {
			t.Fatalf("expected saved markdown to contain %q", c)
		}
	}
}

func TestDeleteSkillRemovesMatchingMarkdownFile(t *testing.T) {
	tmp := t.TempDir()
	s, err := New(tmp)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	_, err = s.SaveSkill(model.Skill{
		Name:    "Delete Me",
		Version: "0.1.0",
		Body:    "# Delete Me\n\nBody",
	})
	if err != nil {
		t.Fatalf("save skill: %v", err)
	}

	deletedPath, err := s.DeleteSkill("delete me")
	if err != nil {
		t.Fatalf("delete skill: %v", err)
	}
	if deletedPath == "" {
		t.Fatal("expected deleted path")
	}

	skills, err := s.ListSkills()
	if err != nil {
		t.Fatalf("list skills: %v", err)
	}
	if len(skills) != 0 {
		t.Fatalf("expected no skills after delete, got %d", len(skills))
	}
}
