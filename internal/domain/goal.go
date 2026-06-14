package domain

import "time"

type Goal struct {
	ID          string    `gorm:"primaryKey"       json:"id"`
	UserID      string    `gorm:"not null;index"   json:"-"`
	Name        string    `gorm:"not null"         json:"name"`
	TargetValue float64   `json:"targetValue"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type GoalRepository interface {
	Create(goal *Goal) error
	FindAll(userID string) ([]Goal, error)
	Update(goal *Goal) error
	Delete(id string) error
}
