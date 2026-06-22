package persistence

import (
	"errors"

	"carteira-inteligente-api/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormAllocationRepository struct {
	db *gorm.DB
}

func NewGormAllocationRepository(db *gorm.DB) domain.AllocationRepository {
	return &GormAllocationRepository{db: db}
}

func (r *GormAllocationRepository) Get(userID string) (*domain.AllocationConfig, error) {
	var cfg domain.AllocationConfig
	err := r.db.First(&cfg, "user_id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *GormAllocationRepository) Upsert(cfg *domain.AllocationConfig) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		UpdateAll: true,
	}).Create(cfg).Error
}
