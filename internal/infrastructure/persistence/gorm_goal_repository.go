package persistence

import (
	"carteira-inteligente-api/internal/domain"

	"gorm.io/gorm"
)

type GormGoalRepository struct {
	db *gorm.DB
}

func NewGormGoalRepository(db *gorm.DB) domain.GoalRepository {
	return &GormGoalRepository{db: db}
}

func (r *GormGoalRepository) Create(goal *domain.Goal) error {
	return r.db.Create(goal).Error
}

func (r *GormGoalRepository) FindAll() ([]domain.Goal, error) {
	var goals []domain.Goal
	if err := r.db.Order("created_at asc").Find(&goals).Error; err != nil {
		return nil, err
	}
	return goals, nil
}

func (r *GormGoalRepository) Update(goal *domain.Goal) error {
	result := r.db.Save(goal)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *GormGoalRepository) Delete(id string) error {
	result := r.db.Delete(&domain.Goal{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
