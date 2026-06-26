package domain

import "time"

// Asset é um ativo do catálogo da B3 (tabela b3_assets). O Type é a classificação
// autoritativa (Acoes | FIIs | ETFs) derivada da categoria de origem no
// Investidor10 — não de heurística por sufixo. É a fonte de verdade que separa
// units de ação terminadas em 11 (TAEE11/SANB11 → Acoes) de FIIs e ETFs.
type Asset struct {
	Ticker    string    `gorm:"primaryKey"   json:"ticker"`
	Name      string    `json:"name"`
	Type      string    `gorm:"index"        json:"type"`
	Sector    string    `json:"sector"`
	UpdatedAt time.Time `json:"-"`
}

// TableName fixa o nome da tabela do catálogo.
func (Asset) TableName() string { return "b3_assets" }

// AssetRepository persiste e consulta o catálogo da B3.
type AssetRepository interface {
	// Upsert insere/atualiza ativos por ticker de forma idempotente.
	Upsert(assets []Asset) error
	// FindByTicker resolve um ticker; ErrNotFound quando ausente do catálogo.
	FindByTicker(ticker string) (*Asset, error)
	// Search retorna ativos por prefixo de ticker (ou nome), limitado a limit.
	Search(prefix string, limit int) ([]Asset, error)
	// Count informa o tamanho atual do catálogo.
	Count() (int64, error)
}
