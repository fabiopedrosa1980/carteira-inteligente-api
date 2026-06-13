package domain

import "time"

type Goal struct {
	ID          string    `gorm:"primaryKey"       json:"id"`
	UserID      string    `gorm:"not null;index"   json:"-"`
	Name        string    `gorm:"not null"         json:"name"`
	TargetValue float64   `json:"targetValue"`
	Type        string    `gorm:"not null"         json:"type"` // patrimonio | renda_mensal | preco_medio
	Ticker      string    `json:"ticker"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type GoalRepository interface {
	Create(goal *Goal) error
	FindAll(userID string) ([]Goal, error)
	Update(goal *Goal) error
	Delete(id string) error
}
