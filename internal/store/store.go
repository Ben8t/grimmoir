package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"grimmoir/internal/model"
	"gopkg.in/yaml.v3"
)

type Store struct {
	root string
}

type frontmatter struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
	Version     string   `yaml:"version"`
}

func New(root string) (*Store, error) {
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("store root is required")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("ensure store root: %w", err)
	}
	return &Store{root: root}, nil
}

func (s *Store) ListSkills() ([]model.Skill, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		return nil, fmt.Errorf("read store root: %w", err)
	}

	skills := make([]model.Skill, 0)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		path := filepath.Join(s.root, entry.Name())
		skill, err := parseSkillFile(path)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}

	sort.Slice(skills, func(i, j int) bool {
		return strings.ToLower(skills[i].Name) < strings.ToLower(skills[j].Name)
	})

	return skills, nil
}

func (s *Store) SaveSkill(skill model.Skill) (string, error) {
	fileName := slugify(skill.Name)
	if fileName == "" {
		return "", errors.New("name is required")
	}

	path := filepath.Join(s.root, fileName+".md")
	fmText := renderFrontmatter(frontmatter{
		Name:        skill.Name,
		Description: skill.Description,
		Tags:        skill.Tags,
		Version:     skill.Version,
	})
	body := strings.TrimSpace(skill.Body)
	content := "---\n" + fmText + "---\n\n" + body + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write skill file: %w", err)
	}

	return path, nil
}

func (s *Store) DeleteSkill(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", errors.New("name is required")
	}
	skills, err := s.ListSkills()
	if err != nil {
		return "", err
	}
	for _, skill := range skills {
		if strings.EqualFold(skill.Name, name) {
			if err := os.Remove(skill.Path); err != nil {
				return "", fmt.Errorf("remove skill file: %w", err)
			}
			return skill.Path, nil
		}
	}
	return "", fmt.Errorf("skill %q not found", name)
}

func renderFrontmatter(fm frontmatter) string {
	tags := "[]"
	if len(fm.Tags) > 0 {
		tags = "[" + strings.Join(fm.Tags, ", ") + "]"
	}
	return fmt.Sprintf(
		"name: %q\ndescription: %q\ntags: %s\nversion: %q\n",
		fm.Name,
		fm.Description,
		tags,
		fm.Version,
	)
}

func parseSkillFile(path string) (model.Skill, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return model.Skill{}, fmt.Errorf("read %s: %w", path, err)
	}

	text := string(b)
	parts := strings.SplitN(text, "---", 3)
	if len(parts) < 3 {
		return model.Skill{}, fmt.Errorf("invalid frontmatter in %s", path)
	}

	var fm frontmatter
	if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
		return model.Skill{}, fmt.Errorf("parse frontmatter in %s: %w", path, err)
	}

	return model.Skill{
		Name:        fm.Name,
		Description: fm.Description,
		Tags:        fm.Tags,
		Version:     fm.Version,
		Body:        strings.TrimSpace(parts[2]),
		Path:        path,
	}, nil
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	b := strings.Builder{}
	lastDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteRune('_')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "_")
	return out
}
