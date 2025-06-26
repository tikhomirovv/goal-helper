package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"goal-helper/internal/models"
)

// OpenAIClient представляет клиент для работы с OpenAI API
type OpenAIClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
	model      string
}

// NewOpenAIClient создает новый OpenAI клиент
func NewOpenAIClient(apiKey string) Client {
	return &OpenAIClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.openai.com/v1/chat/completions",
		model:   "gpt-4o-mini-2024-07-18", // Можно сделать настраиваемым
	}
}

// GenerateStep генерирует следующий шаг для цели
func (c *OpenAIClient) GenerateStep(goal *models.Goal, completedSteps []*models.Step) (*StepResponse, error) {
	prompt := c.buildStepPrompt(goal, completedSteps)

	response, err := c.callOpenAI(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI: %w", err)
	}

	var stepResponse StepResponse
	if err := json.Unmarshal([]byte(response), &stepResponse); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	return &stepResponse, nil
}

// RephraseStep переформулирует текущий шаг
func (c *OpenAIClient) RephraseStep(goal *models.Goal, currentStep *models.Step, userComment string) (*StepResponse, error) {
	prompt := c.buildRephrasePrompt(goal, currentStep, userComment)

	response, err := c.callOpenAI(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI: %w", err)
	}

	var stepResponse StepResponse
	if err := json.Unmarshal([]byte(response), &stepResponse); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	return &stepResponse, nil
}

// ClarifyGoal запрашивает уточнение цели
func (c *OpenAIClient) ClarifyGoal(goalTitle, goalDescription string) (*ClarificationResponse, error) {
	prompt := c.buildClarificationPrompt(goalTitle, goalDescription)

	response, err := c.callOpenAI(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI: %w", err)
	}

	var clarificationResponse ClarificationResponse
	if err := json.Unmarshal([]byte(response), &clarificationResponse); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	return &clarificationResponse, nil
}

// GenerateGoalTitle генерирует название цели на основе описания
func (c *OpenAIClient) GenerateGoalTitle(description string) (string, error) {
	prompt := c.buildTitlePrompt(description)
	log.Printf("🔍 Prompt для генерации названия цели: %s", prompt)

	response, err := c.callOpenAI(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI: %w", err)
	}
	var titleResponse struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal([]byte(response), &titleResponse); err != nil {
		return "", fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	return titleResponse.Title, nil
}

// callOpenAI отправляет запрос к OpenAI API
func (c *OpenAIClient) callOpenAI(prompt string) (string, error) {
	// Если API ключ не установлен, используем заглушку для тестирования
	if c.apiKey == "" {
		return c.mockResponse(prompt)
	}

	// Структура запроса к OpenAI
	requestBody := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "Ты помощник для достижения целей. Всегда отвечай в формате JSON.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
		"max_tokens":  500,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Парсим ответ OpenAI
	var openAIResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openAIResponse); err != nil {
		return "", fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(openAIResponse.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenAI response")
	}

	return openAIResponse.Choices[0].Message.Content, nil
}

// mockResponse возвращает заглушку для тестирования без API ключа
func (c *OpenAIClient) mockResponse(prompt string) (string, error) {
	if strings.Contains(prompt, "уточн") {
		return `{"status": "need_clarification", "question": "У тебя уже есть конкретная идея для этой цели?"}`, nil
	}

	if strings.Contains(prompt, "название") {
		return `{"title": "Достичь поставленной цели"}`, nil
	}

	return `{"status": "ok", "step": "Начни с составления плана на бумаге"}`, nil
}

// buildStepPrompt формирует промпт для генерации шага
func (c *OpenAIClient) buildStepPrompt(goal *models.Goal, completedSteps []*models.Step) string {
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
func (c *OpenAIClient) buildRephrasePrompt(goal *models.Goal, currentStep *models.Step, userComment string) string {
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
func (c *OpenAIClient) buildClarificationPrompt(goalTitle, goalDescription string) string {
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

// buildTitlePrompt формирует промпт для генерации названия цели
func (c *OpenAIClient) buildTitlePrompt(description string) string {
	var prompt strings.Builder

	prompt.WriteString("Сгенерируй краткое и точное название для цели на основе описания.\n\n")
	prompt.WriteString("Описание: " + description + "\n\n")
	prompt.WriteString("Название должно быть:\n")
	prompt.WriteString("- Кратким (3-7 слов)\n")
	prompt.WriteString("- Конкретным и понятным\n")
	prompt.WriteString("- Мотивирующим\n")
	prompt.WriteString("- Без кавычек\n\n")
	prompt.WriteString("Формат ответа (JSON):\n")
	prompt.WriteString(`{
  "title": "краткое название цели"
}`)

	return prompt.String()
}
