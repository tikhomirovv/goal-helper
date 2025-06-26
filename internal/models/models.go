package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// User представляет пользователя Telegram
type User struct {
	ID           string    `json:"id"`                       // Telegram User ID
	Username     string    `json:"username"`                 // Telegram username
	FirstName    string    `json:"first_name"`               // Имя пользователя
	CreatedAt    time.Time `json:"created_at"`               // Дата создания
	ActiveGoalID string    `json:"active_goal_id,omitempty"` // ID активной цели
}

// Goal представляет цель пользователя
type Goal struct {
	ID          string     `json:"id"`                     // Уникальный ID цели
	UserID      string     `json:"user_id"`                // ID пользователя
	Title       string     `json:"title"`                  // Название цели
	Description string     `json:"description"`            // Описание цели
	CreatedAt   time.Time  `json:"created_at"`             // Дата создания
	UpdatedAt   time.Time  `json:"updated_at"`             // Дата обновления
	Context     Context    `json:"context"`                // Контекст для LLM
	Status      string     `json:"status"`                 // "active", "completed", "abandoned"
	CompletedAt *time.Time `json:"completed_at,omitempty"` // Дата завершения
}

// Step представляет шаг к достижению цели
type Step struct {
	ID          string     `json:"id"`                     // Уникальный ID шага
	GoalID      string     `json:"goal_id"`                // ID цели
	Text        string     `json:"text"`                   // Текст шага
	CreatedAt   time.Time  `json:"created_at"`             // Дата создания
	CompletedAt *time.Time `json:"completed_at,omitempty"` // Дата выполнения
	Rephrased   bool       `json:"rephrased"`              // Был ли переформулирован
	UserComment string     `json:"user_comment,omitempty"` // Комментарий пользователя
}

// Context содержит дополнительную информацию для LLM
type Context struct {
	Clarifications []string `json:"clarifications"`  // Уточняющие вопросы и ответы
	Notes          string   `json:"notes,omitempty"` // Дополнительные заметки
}

// NewUser создает нового пользователя
func NewUser(id, username, firstName string) *User {
	return &User{
		ID:        id,
		Username:  username,
		FirstName: firstName,
		CreatedAt: time.Now(),
	}
}

// NewGoal создает новую цель
func NewGoal(userID, title, description string) *Goal {
	return &Goal{
		ID:          uuid.New().String(),
		UserID:      userID,
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Context:     Context{Clarifications: []string{}},
		Status:      "active",
	}
}

// NewStep создает новый шаг
func NewStep(goalID, text string) *Step {
	return &Step{
		ID:        uuid.New().String(),
		GoalID:    goalID,
		Text:      text,
		CreatedAt: time.Now(),
		Rephrased: false,
	}
}

// IsCompleted проверяет, выполнен ли шаг
func (s *Step) IsCompleted() bool {
	return s.CompletedAt != nil
}

// Complete отмечает шаг как выполненный
func (s *Step) Complete() {
	now := time.Now()
	s.CompletedAt = &now
}

// Rephrase отмечает шаг как переформулированный
func (s *Step) Rephrase(comment string) {
	s.Rephrased = true
	s.UserComment = comment
}

// AddClarification добавляет уточнение в контекст цели
func (g *Goal) AddClarification(question, answer string) {
	clarification := fmt.Sprintf("Вопрос: %s | Ответ: %s", question, answer)
	g.Context.Clarifications = append(g.Context.Clarifications, clarification)
	g.UpdatedAt = time.Now()
}

// GetContextSummary возвращает краткое описание собранного контекста
func (g *Goal) GetContextSummary() string {
	if len(g.Context.Clarifications) == 0 {
		return "Контекст не собран"
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Собрано %d уточнений:\n", len(g.Context.Clarifications)))
	for i, clarification := range g.Context.Clarifications {
		summary.WriteString(fmt.Sprintf("%d. %s\n", i+1, clarification))
	}
	return summary.String()
}
