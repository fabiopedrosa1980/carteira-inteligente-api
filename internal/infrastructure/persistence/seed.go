package persistence

import (
	"carteira-inteligente-api/internal/domain"

	"gorm.io/gorm"
)

func SeedIfEmpty(db *gorm.DB) error {
	var count int64
	db.Model(&domain.Stock{}).Count(&count)
	if count > 0 {
		return nil
	}

	stocks := []domain.Stock{
		{Ticker: "BBAS3", Nome: "Banco do Brasil", Setor: "Bancário", Nota: 8, PrecoAtual: 55.2, VariacaoHoje: 0, DY: 8.5},
		{Ticker: "BBSE3", Nome: "BB Seguridade", Setor: "Seguros", Nota: 9, PrecoAtual: 34.8, VariacaoHoje: 0, DY: 9.2},
		{Ticker: "PETR4", Nome: "Petrobras", Setor: "Petróleo & Gás", Nota: 9, PrecoAtual: 37.9, VariacaoHoje: 0, DY: 15.3},
		{Ticker: "ITUB3", Nome: "Itaú Unibanco", Setor: "Bancário", Nota: 4, PrecoAtual: 35.1, VariacaoHoje: 0, DY: 3.5},
		{Ticker: "BRAP4", Nome: "Bradespar", Setor: "Mineração", Nota: 9, PrecoAtual: 17.8, VariacaoHoje: 0, DY: 10.1},
		{Ticker: "CMIG4", Nome: "Cemig", Setor: "Energia Elétrica", Nota: 9, PrecoAtual: 11.95, VariacaoHoje: 0, DY: 10.8},
		{Ticker: "CPFE3", Nome: "CPFL Energia", Setor: "Energia Elétrica", Nota: 7, PrecoAtual: 35.5, VariacaoHoje: 0, DY: 7.2},
		{Ticker: "CSMG3", Nome: "Copasa", Setor: "Saneamento", Nota: 8, PrecoAtual: 20.4, VariacaoHoje: 0, DY: 8.1},
		{Ticker: "ISAE4", Nome: "Isa Cteep", Setor: "Energia Elétrica", Nota: 9, PrecoAtual: 25.3, VariacaoHoje: 0, DY: 9.6},
		{Ticker: "CXSE3", Nome: "Caixa Seguridade", Setor: "Seguros", Nota: 8, PrecoAtual: 14.9, VariacaoHoje: 0, DY: 8.3},
	}

	for i := range stocks {
		if err := db.Create(&stocks[i]).Error; err != nil {
			return err
		}
	}

	type entry struct {
		ticker string
		months []int
		amount float64
		typ    domain.DividendType
	}

	// yearSeed define os pagamentos de cada ação por ano.
	// Mesmos meses do seed 2025; amounts derivados do histórico real (Investidor10).
	// BBAS3/ITUB3/CSMG3/ISAE4 → JCP predominante; demais → dividendo.
	type yearSeed struct {
		year      int
		dividends []entry
	}

	seeds := []yearSeed{
		{
			year: 2021,
			dividends: []entry{
				{ticker: "BBAS3", months: []int{1, 4, 6, 9, 12}, amount: 0.12, typ: domain.DividendTypeJCP},
				{ticker: "BBSE3", months: []int{1, 3, 6, 8, 11}, amount: 0.20, typ: domain.DividendTypeDividendo},
				{ticker: "PETR4", months: []int{2, 5, 8, 11}, amount: 0.60, typ: domain.DividendTypeDividendo},
				{ticker: "ITUB3", months: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, amount: 0.05, typ: domain.DividendTypeJCP},
				{ticker: "BRAP4", months: []int{3, 6, 9, 12}, amount: 1.96, typ: domain.DividendTypeDividendo},
				{ticker: "CMIG4", months: []int{3, 6, 9, 12}, amount: 0.22, typ: domain.DividendTypeDividendo},
				{ticker: "CPFE3", months: []int{2, 5, 8, 11}, amount: 0.75, typ: domain.DividendTypeDividendo},
				{ticker: "CSMG3", months: []int{4, 8, 12}, amount: 0.25, typ: domain.DividendTypeJCP},
				{ticker: "ISAE4", months: []int{3, 6, 9, 12}, amount: 0.37, typ: domain.DividendTypeJCP},
				{ticker: "CXSE3", months: []int{3, 6, 9, 12}, amount: 0.12, typ: domain.DividendTypeDividendo},
			},
		},
		{
			year: 2022,
			dividends: []entry{
				{ticker: "BBAS3", months: []int{1, 4, 6, 9, 12}, amount: 0.21, typ: domain.DividendTypeJCP},
				{ticker: "BBSE3", months: []int{1, 3, 6, 8, 11}, amount: 0.39, typ: domain.DividendTypeDividendo},
				{ticker: "PETR4", months: []int{2, 5, 8, 11}, amount: 2.93, typ: domain.DividendTypeDividendo},
				{ticker: "ITUB3", months: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, amount: 0.05, typ: domain.DividendTypeJCP},
				{ticker: "BRAP4", months: []int{3, 6, 9, 12}, amount: 0.78, typ: domain.DividendTypeDividendo},
				{ticker: "CMIG4", months: []int{3, 6, 9, 12}, amount: 0.19, typ: domain.DividendTypeDividendo},
				{ticker: "CPFE3", months: []int{2, 5, 8, 11}, amount: 0.81, typ: domain.DividendTypeDividendo},
				{ticker: "CSMG3", months: []int{4, 8, 12}, amount: 0.13, typ: domain.DividendTypeJCP},
				{ticker: "ISAE4", months: []int{3, 6, 9, 12}, amount: 0.27, typ: domain.DividendTypeJCP},
				{ticker: "CXSE3", months: []int{3, 6, 9, 12}, amount: 0.16, typ: domain.DividendTypeDividendo},
			},
		},
		{
			year: 2023,
			dividends: []entry{
				{ticker: "BBAS3", months: []int{1, 4, 6, 9, 12}, amount: 0.23, typ: domain.DividendTypeJCP},
				{ticker: "BBSE3", months: []int{1, 3, 6, 8, 11}, amount: 0.69, typ: domain.DividendTypeDividendo},
				{ticker: "PETR4", months: []int{2, 5, 8, 11}, amount: 1.53, typ: domain.DividendTypeDividendo},
				{ticker: "ITUB3", months: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, amount: 0.05, typ: domain.DividendTypeJCP},
				{ticker: "BRAP4", months: []int{3, 6, 9, 12}, amount: 0.80, typ: domain.DividendTypeDividendo},
				{ticker: "CMIG4", months: []int{3, 6, 9, 12}, amount: 0.08, typ: domain.DividendTypeDividendo},
				{ticker: "CPFE3", months: []int{2, 5, 8, 11}, amount: 0.53, typ: domain.DividendTypeDividendo},
				{ticker: "CSMG3", months: []int{4, 8, 12}, amount: 0.56, typ: domain.DividendTypeJCP},
				{ticker: "ISAE4", months: []int{3, 6, 9, 12}, amount: 0.61, typ: domain.DividendTypeJCP},
				{ticker: "CXSE3", months: []int{3, 6, 9, 12}, amount: 0.25, typ: domain.DividendTypeDividendo},
			},
		},
		{
			year: 2024,
			dividends: []entry{
				{ticker: "BBAS3", months: []int{1, 4, 6, 9, 12}, amount: 0.27, typ: domain.DividendTypeJCP},
				{ticker: "BBSE3", months: []int{1, 3, 6, 8, 11}, amount: 0.53, typ: domain.DividendTypeDividendo},
				{ticker: "PETR4", months: []int{2, 5, 8, 11}, amount: 1.29, typ: domain.DividendTypeDividendo},
				{ticker: "ITUB3", months: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, amount: 0.12, typ: domain.DividendTypeJCP},
				{ticker: "BRAP4", months: []int{3, 6, 9, 12}, amount: 0.52, typ: domain.DividendTypeDividendo},
				{ticker: "CMIG4", months: []int{3, 6, 9, 12}, amount: 0.08, typ: domain.DividendTypeDividendo},
				{ticker: "CPFE3", months: []int{2, 5, 8, 11}, amount: 0.69, typ: domain.DividendTypeDividendo},
				{ticker: "CSMG3", months: []int{4, 8, 12}, amount: 0.69, typ: domain.DividendTypeJCP},
				{ticker: "ISAE4", months: []int{3, 6, 9, 12}, amount: 0.55, typ: domain.DividendTypeJCP},
				{ticker: "CXSE3", months: []int{3, 6, 9, 12}, amount: 0.27, typ: domain.DividendTypeDividendo},
			},
		},
		{
			year: 2025,
			dividends: []entry{
				{ticker: "BBAS3", months: []int{1, 4, 6, 9, 12}, amount: 0.39, typ: domain.DividendTypeJCP},
				{ticker: "BBSE3", months: []int{1, 3, 6, 8, 11}, amount: 0.27, typ: domain.DividendTypeDividendo},
				{ticker: "PETR4", months: []int{2, 5, 8, 11}, amount: 0.49, typ: domain.DividendTypeDividendo},
				{ticker: "ITUB3", months: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, amount: 0.10, typ: domain.DividendTypeJCP},
				{ticker: "BRAP4", months: []int{3, 6, 9, 12}, amount: 0.15, typ: domain.DividendTypeDividendo},
				{ticker: "CMIG4", months: []int{3, 6, 9, 12}, amount: 0.11, typ: domain.DividendTypeDividendo},
				{ticker: "CPFE3", months: []int{2, 5, 8, 11}, amount: 0.21, typ: domain.DividendTypeDividendo},
				{ticker: "CSMG3", months: []int{4, 8, 12}, amount: 0.14, typ: domain.DividendTypeJCP},
				{ticker: "ISAE4", months: []int{3, 6, 9, 12}, amount: 0.20, typ: domain.DividendTypeJCP},
				{ticker: "CXSE3", months: []int{3, 6, 9, 12}, amount: 0.10, typ: domain.DividendTypeDividendo},
			},
		},
	}

	tickerToID := make(map[string]uint, len(stocks))
	for _, s := range stocks {
		tickerToID[s.Ticker] = s.ID
	}

	for _, ys := range seeds {
		for _, e := range ys.dividends {
			id, ok := tickerToID[e.ticker]
			if !ok {
				continue
			}
			for _, m := range e.months {
				d := domain.Dividend{
					StockID: id,
					Amount:  e.amount,
					Month:   m,
					Year:    ys.year,
					Type:    e.typ,
				}
				if err := db.Create(&d).Error; err != nil {
					return err
				}
			}
		}
	}

	return nil
}
