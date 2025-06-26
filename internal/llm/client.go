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
}

// StepResponse представляет ответ LLM на генерацию шага
type StepResponse struct {
	Status           string `json:"status"` // "ok", "need_clarification", "goal_completed", "near_completion"
	Step             string `json:"step,omitempty"`
	Question         string `json:"question,omitempty"`
	CompletionReason string `json:"completion_reason,omitempty"` // Причина завершения цели
}

// ClarificationResponse представляет ответ LLM на уточнение цели
type ClarificationResponse struct {
	Status   string `json:"status"` // "ok" или "need_clarification"
	Question string `json:"question,omitempty"`
}
