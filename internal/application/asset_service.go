package application

import (
	"log"
	"time"

	"carteira-inteligente-api/internal/domain"
	"carteira-inteligente-api/internal/infrastructure/scraper"
)

// AssetUseCase expõe a resolução do catálogo da B3 e o seu refresh.
type AssetUseCase interface {
	GetByTicker(ticker string) (*domain.Asset, error)
	Search(prefix string, limit int) ([]domain.Asset, error)
	Refresh() (int, error)
}

type AssetService struct {
	repo domain.AssetRepository
}

func NewAssetService(repo domain.AssetRepository) *AssetService {
	return &AssetService{repo: repo}
}

func (s *AssetService) GetByTicker(ticker string) (*domain.Asset, error) {
	return s.repo.FindByTicker(ticker)
}

func (s *AssetService) Search(prefix string, limit int) ([]domain.Asset, error) {
	if limit <= 0 || limit > 50 {
		limit = 15
	}
	return s.repo.Search(prefix, limit)
}

// Refresh coleta o catálogo completo da B3 (sitemaps do Investidor10) e faz
// upsert. Em falha de download retorna erro sem tocar o catálogo já persistido.
func (s *AssetService) Refresh() (int, error) {
	assets, err := scraper.FetchB3Catalog()
	if err != nil {
		return 0, err
	}
	now := time.Now()
	for i := range assets {
		assets[i].UpdatedAt = now
	}
	if err := s.repo.Upsert(assets); err != nil {
		return 0, err
	}
	return len(assets), nil
}

// StartCatalogSync popula o catálogo no startup e o atualiza periodicamente.
// Best-effort: falha de uma execução não derruba a API — o catálogo da última
// ingestão bem-sucedida permanece servindo. Pula o refresh inicial quando já há
// catálogo persistido (evita custo no cold start do Render).
func StartCatalogSync(svc *AssetService, interval time.Duration) {
	go func() {
		if n, err := svc.repo.Count(); err == nil && n > 0 {
			log.Printf("[catalog] %d ativos já persistidos; aguardando próximo ciclo", n)
			time.Sleep(interval)
		}
		for {
			if n, err := svc.Refresh(); err != nil {
				log.Printf("[catalog] refresh: %v", err)
			} else {
				log.Printf("[catalog] refresh: %d ativos no catálogo", n)
			}
			time.Sleep(interval)
		}
	}()
}
