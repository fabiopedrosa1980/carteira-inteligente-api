package application

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"carteira-inteligente-api/internal/domain"
)

type GoalService struct {
	repo domain.GoalRepository
}

func NewGoalService(repo domain.GoalRepository) *GoalService {
	return &GoalService{repo: repo}
}

func (s *GoalService) ListGoals(userID string) ([]domain.Goal, error) {
	return s.repo.FindAll(userID)
}

func (s *GoalService) CreateGoal(userID string, g *domain.Goal) error {
	if strings.TrimSpace(g.Name) == "" {
		return domain.ErrValidation
	}
	g.ID = newUUID()
	g.UserID = userID
	g.CreatedAt = time.Now()
	g.UpdatedAt = time.Now()
	return s.repo.Create(g)
}

func (s *GoalService) UpdateGoal(userID, id string, updated *domain.Goal) (*domain.Goal, error) {
	goals, err := s.repo.FindAll(userID)
	if err != nil {
		return nil, err
	}
	var existing *domain.Goal
	for i := range goals {
		if goals[i].ID == id {
			existing = &goals[i]
			break
		}
	}
	if existing == nil {
		return nil, domain.ErrNotFound
	}
	existing.Name = updated.Name
	existing.TargetValue = updated.TargetValue
	existing.UpdatedAt = time.Now()
	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *GoalService) DeleteGoal(id string) error {
	return s.repo.Delete(id)
}

func newUUID() string {
	var b [16]byte
	rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
