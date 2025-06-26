package llm

import (
	"goal-helper/internal/models"
)

// Client представляет интерфейс для работы с LLM
type Client interface {
	GenerateStep(goal *models.Goal, completedSteps []*models.Step) (*StepResponse, error)
	RephraseStep(goal *models.Goal, currentStep *models.Step, userComment string) (*StepResponse, error)
	ClarifyGoal(goalTitle, goalDescription string) (*ClarificationResponse, error)
	GenerateGoalTitle(description string) (string, error)
	GatherContext(goal *models.Goal) (*ContextResponse, error)
}

// StepResponse представляет ответ LLM на генерацию шага
type StepResponse struct {
	Status           string `json:"status"`            // "ok", "need_clarification", "goal_completed", "near_completion"
	Step             string `json:"step"`              // Текст шага (может быть пустым, если нужна дополнительная информация)
	Question         string `json:"question"`          // Уточняющий вопрос (может быть пустым, если шаг понятен)
	CompletionReason string `json:"completion_reason"` // Причина завершения цели (может быть пустым, если цель не завершена)
}

// ClarificationResponse представляет ответ LLM на уточнение цели
type ClarificationResponse struct {
	Status   string `json:"status"`   // "need_clarification"
	Question string `json:"question"` // Уточняющий вопрос
}

// ContextResponse представляет ответ LLM на сбор контекста
type ContextResponse struct {
	Status   string `json:"status"`   // "ok" или "need_context"
	Question string `json:"question"` // Вопрос для сбора контекста (может быть пустым, если контекст собран)
	Context  string `json:"context"`  // Собранный контекст (может быть пустым, если нужен дополнительный контекст)
}
