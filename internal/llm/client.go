package llm

import (
	"encoding/json"
	"fmt"
	"goal-helper/internal/models"
	"net/http"
	"strings"
	"time"
)

// Client представляет интерфейс для работы с LLM
type Client interface {
	GenerateStep(goal *models.Goal, completedSteps []*models.Step) (*StepResponse, error)
	RephraseStep(goal *models.Goal, currentStep *models.Step, userComment string) (*StepResponse, error)
	ClarifyGoal(goalTitle, goalDescription string) (*ClarificationResponse, error)
}

// StepResponse представляет ответ LLM на генерацию шага
type StepResponse struct {
	Status   string `json:"status"` // "ok" или "need_clarification"
	Step     string `json:"step,omitempty"`
	Question string `json:"question,omitempty"`
}

// ClarificationResponse представляет ответ LLM на уточнение цели
type ClarificationResponse struct {
	Status   string `json:"status"` // "ok" или "need_clarification"
	Question string `json:"question,omitempty"`
}

// HTTPClient представляет HTTP клиент для работы с LLM API
type HTTPClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewClient создает новый LLM клиент
func NewClient(apiKey string) Client {
	return &HTTPClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.openai.com/v1/chat/completions", // По умолчанию OpenAI
	}
}

// GenerateStep генерирует следующий шаг для цели
func (c *HTTPClient) GenerateStep(goal *models.Goal, completedSteps []*models.Step) (*StepResponse, error) {
	// Формируем промпт для LLM
	prompt := c.buildStepPrompt(goal, completedSteps)

	// Отправляем запрос к LLM
	response, err := c.callLLM(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	// Парсим ответ
	var stepResponse StepResponse
	if err := json.Unmarshal([]byte(response), &stepResponse); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return &stepResponse, nil
}

// RephraseStep переформулирует текущий шаг
func (c *HTTPClient) RephraseStep(goal *models.Goal, currentStep *models.Step, userComment string) (*StepResponse, error) {
	// Формируем промпт для переформулировки
	prompt := c.buildRephrasePrompt(goal, currentStep, userComment)

	// Отправляем запрос к LLM
	response, err := c.callLLM(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	// Парсим ответ
	var stepResponse StepResponse
	if err := json.Unmarshal([]byte(response), &stepResponse); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return &stepResponse, nil
}

// ClarifyGoal запрашивает уточнение цели
func (c *HTTPClient) ClarifyGoal(goalTitle, goalDescription string) (*ClarificationResponse, error) {
	// Формируем промпт для уточнения
	prompt := c.buildClarificationPrompt(goalTitle, goalDescription)

	// Отправляем запрос к LLM
	response, err := c.callLLM(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	// Парсим ответ
	var clarificationResponse ClarificationResponse
	if err := json.Unmarshal([]byte(response), &clarificationResponse); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return &clarificationResponse, nil
}

// buildStepPrompt формирует промпт для генерации шага
func (c *HTTPClient) buildStepPrompt(goal *models.Goal, completedSteps []*models.Step) string {
	var prompt strings.Builder

	prompt.WriteString("Ты коуч, помогаешь пользователю достичь цели, разбивая её на минимальные, простые задачи.\n\n")
	prompt.WriteString("Цель: " + goal.Title + "\n")
	if goal.Description != "" {
		prompt.WriteString("Описание: " + goal.Description + "\n")
	}

	if len(completedSteps) > 0 {
		prompt.WriteString("Выполненные шаги:\n")
		for i, step := range completedSteps {
			prompt.WriteString(fmt.Sprintf("%d. %s\n", i+1, step.Text))
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("Сгенерируй следующий логичный шаг. Если не уверен — задай вопрос.\n\n")
	prompt.WriteString("Формат ответа (JSON):\n")
	prompt.WriteString(`{
  "status": "ok" | "need_clarification",
  "step": "текст шага",
  "question": "уточняющий вопрос (если нужен)"
}`)

	return prompt.String()
}

// buildRephrasePrompt формирует промпт для переформулировки шага
func (c *HTTPClient) buildRephrasePrompt(goal *models.Goal, currentStep *models.Step, userComment string) string {
	var prompt strings.Builder

	prompt.WriteString("Цель: " + goal.Title + "\n")
	prompt.WriteString("Текущий шаг: " + currentStep.Text + "\n")
	prompt.WriteString("Комментарий пользователя: " + userComment + "\n\n")
	prompt.WriteString("Сформулируй альтернативный шаг на том же уровне сложности.\n\n")
	prompt.WriteString("Формат ответа (JSON):\n")
	prompt.WriteString(`{
  "status": "ok",
  "step": "новый текст шага"
}`)

	return prompt.String()
}

// buildClarificationPrompt формирует промпт для уточнения цели
func (c *HTTPClient) buildClarificationPrompt(goalTitle, goalDescription string) string {
	var prompt strings.Builder

	prompt.WriteString("Цель: " + goalTitle + "\n")
	if goalDescription != "" {
		prompt.WriteString("Описание: " + goalDescription + "\n")
	}
	prompt.WriteString("\nЕсли цель недостаточно понятна для генерации шага — верни статус и вопрос:\n\n")
	prompt.WriteString("Формат ответа (JSON):\n")
	prompt.WriteString(`{
  "status": "need_clarification",
  "question": "уточняющий вопрос"
}`)

	return prompt.String()
}

// callLLM отправляет запрос к LLM API
func (c *HTTPClient) callLLM(prompt string) (string, error) {
	// TODO: Реализовать вызов конкретного LLM API
	// Пока возвращаем заглушку для тестирования

	// Простая заглушка для тестирования
	if strings.Contains(prompt, "уточн") {
		return `{"status": "need_clarification", "question": "У тебя уже есть конкретная идея для этой цели?"}`, nil
	}

	return `{"status": "ok", "step": "Начни с составления плана на бумаге"}`, nil
}
