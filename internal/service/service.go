package service

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"grimmoir/internal/model"
)

type Repository interface {
	ListSkills() ([]model.Skill, error)
	SaveSkill(skill model.Skill) (string, error)
	DeleteSkill(name string) (string, error)
}

type Clipboard interface {
	ReadAll() (string, error)
	WriteAll(text string) error
}

type Service struct {
	repo      Repository
	clipboard Clipboard
}

func New(repo Repository, clipboard Clipboard) *Service {
	return &Service{repo: repo, clipboard: clipboard}
}

func (s *Service) List() ([]model.Skill, error) {
	skills, err := s.repo.ListSkills()
	if err != nil {
		return nil, fmt.Errorf("list skills: %w", err)
	}
	return skills, nil
}

func (s *Service) GetByName(name string) (model.Skill, error) {
	if strings.TrimSpace(name) == "" {
		return model.Skill{}, errors.New("name is required")
	}
	skills, err := s.List()
	if err != nil {
		return model.Skill{}, err
	}
	for _, skill := range skills {
		if strings.EqualFold(skill.Name, name) {
			return skill, nil
		}
	}
	return model.Skill{}, fmt.Errorf("skill %q not found", name)
}

func (s *Service) AddFromText(skill model.Skill) (string, error) {
	if strings.TrimSpace(skill.Name) == "" {
		return "", errors.New("name is required")
	}
	if strings.TrimSpace(skill.Body) == "" {
		return "", errors.New("body is required")
	}
	if strings.TrimSpace(skill.Version) == "" {
		skill.Version = "0.1.0"
	}
	if strings.TrimSpace(skill.Body) != "" && !strings.HasPrefix(strings.TrimSpace(skill.Body), "#") {
		skill.Body = fmt.Sprintf("# %s\n\n%s", skill.Name, strings.TrimSpace(skill.Body))
	}
	path, err := s.repo.SaveSkill(skill)
	if err != nil {
		return "", fmt.Errorf("save skill: %w", err)
	}
	return path, nil
}

func (s *Service) AddFromClipboard(skill model.Skill) (string, error) {
	if s.clipboard == nil {
		return "", errors.New("clipboard is not configured")
	}
	text, err := s.clipboard.ReadAll()
	if err != nil {
		return "", fmt.Errorf("read clipboard: %w", err)
	}
	skill.Body = text
	return s.AddFromText(skill)
}

func (s *Service) Compose(names []string) (string, error) {
	if len(names) == 0 {
		return "", errors.New("at least one skill is required")
	}
	skills, err := s.List()
	if err != nil {
		return "", err
	}
	index := make(map[string]model.Skill, len(skills))
	for _, skill := range skills {
		index[strings.ToLower(skill.Name)] = skill
	}

	out := make([]string, 0, len(names))
	for _, name := range names {
		skill, ok := index[strings.ToLower(name)]
		if !ok {
			return "", fmt.Errorf("skill %q not found", name)
		}
		body := strings.TrimSpace(skill.Body)
		if strings.HasPrefix(body, "# ") {
			out = append(out, body)
		} else {
			out = append(out, fmt.Sprintf("# %s\n\n%s", skill.Name, body))
		}
	}

	return strings.Join(out, "\n\n"), nil
}

func Search(skills []model.Skill, query string) []model.Skill {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return skills
	}
	filtered := make([]model.Skill, 0)
	for _, skill := range skills {
		haystack := strings.ToLower(skill.Name + " " + skill.Description + " " + strings.Join(skill.Tags, " "))
		if strings.Contains(haystack, q) {
			filtered = append(filtered, skill)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		return len(filtered[i].Name) < len(filtered[j].Name)
	})
	return filtered
}

func (s *Service) CopyToClipboard(text string) error {
	if s.clipboard == nil {
		return errors.New("clipboard is not configured")
	}
	if err := s.clipboard.WriteAll(text); err != nil {
		return fmt.Errorf("write clipboard: %w", err)
	}
	return nil
}

func (s *Service) DeleteByName(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", errors.New("name is required")
	}
	path, err := s.repo.DeleteSkill(name)
	if err != nil {
		return "", fmt.Errorf("delete skill: %w", err)
	}
	return path, nil
}
