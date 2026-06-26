package persistence

import (
	"errors"
	"strings"

	"carteira-inteligente-api/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormAssetRepository struct {
	db *gorm.DB
}

func NewGormAssetRepository(db *gorm.DB) domain.AssetRepository {
	return &GormAssetRepository{db: db}
}

// Upsert grava o catálogo por ticker (idempotente): reexecuções não duplicam e
// atualizam name/type/sector quando mudarem. Escreve em lotes para o catálogo
// completo (~1600 ativos) cair em poucas instruções.
func (r *GormAssetRepository) Upsert(assets []domain.Asset) error {
	if len(assets) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "ticker"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "type", "sector", "updated_at"}),
	}).CreateInBatches(assets, 200).Error
}

func (r *GormAssetRepository) FindByTicker(ticker string) (*domain.Asset, error) {
	var a domain.Asset
	err := r.db.Where("ticker = ?", strings.ToUpper(strings.TrimSpace(ticker))).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// Search casa por prefixo de ticker e, secundariamente, por nome contido. UPPER
// no nome mantém a consulta compatível com SQLite e PostgreSQL.
func (r *GormAssetRepository) Search(prefix string, limit int) ([]domain.Asset, error) {
	p := strings.ToUpper(strings.TrimSpace(prefix))
	if p == "" {
		return []domain.Asset{}, nil
	}
	out := []domain.Asset{}
	if err := r.db.
		Where("ticker LIKE ? OR UPPER(name) LIKE ?", p+"%", "%"+p+"%").
		Order("ticker asc").
		Limit(limit).
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *GormAssetRepository) Count() (int64, error) {
	var n int64
	err := r.db.Model(&domain.Asset{}).Count(&n).Error
	return n, err
}
