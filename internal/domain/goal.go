package domain

import "time"

type Goal struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null"   json:"name"`
	Description string    `json:"description"`
	TargetValue float64   `json:"targetValue"`
	Type        string    `gorm:"not null"   json:"type"` // patrimonio | renda_mensal | preco_medio
	Ticker      string    `json:"ticker"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type GoalRepository interface {
	Create(goal *Goal) error
	FindAll() ([]Goal, error)
	Update(goal *Goal) error
	Delete(id string) error
}
