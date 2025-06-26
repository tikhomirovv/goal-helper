package repository

import (
	"goal-helper/internal/models"
)

// Repository определяет интерфейс для работы с данными
// Это позволяет легко заменить файловое хранение на БД
type Repository interface {
	// Пользователи
	GetUser(userID string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error

	// Цели
	GetGoal(goalID string) (*models.Goal, error)
	GetUserGoals(userID string) ([]*models.Goal, error)
	CreateGoal(goal *models.Goal) error
	UpdateGoal(goal *models.Goal) error
	DeleteGoal(goalID string) error

	// Шаги
	GetStep(stepID string) (*models.Step, error)
	GetGoalSteps(goalID string) ([]*models.Step, error)
	GetCurrentStep(goalID string) (*models.Step, error)
	CreateStep(step *models.Step) error
	UpdateStep(step *models.Step) error
	DeleteStep(stepID string) error

	// Утилиты
	Close() error
}
