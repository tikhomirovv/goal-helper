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

	log.Printf("🔍 Отправляем запрос к OpenAI для цели: %s", goal.Title)
	log.Printf("🔍 Длина промпта: %d символов", len(prompt))

	response, err := c.callOpenAI(prompt)
	if err != nil {
		log.Printf("❌ Ошибка при вызове OpenAI: %v", err)
		return nil, fmt.Errorf("failed to call OpenAI: %w", err)
	}

	log.Printf("🔍 Получен ответ от OpenAI: %s", response)

	var stepResponse StepResponse
	if err := json.Unmarshal([]byte(response), &stepResponse); err != nil {
		log.Printf("❌ Ошибка при парсинге ответа OpenAI: %v", err)
		log.Printf("🔍 Сырой ответ: %s", response)

		// Пытаемся найти JSON в ответе
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}")

		if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
			jsonPart := response[jsonStart : jsonEnd+1]
			log.Printf("🔍 Попытка парсинга найденного JSON: %s", jsonPart)

			if err := json.Unmarshal([]byte(jsonPart), &stepResponse); err != nil {
				log.Printf("❌ Ошибка при парсинге найденного JSON: %v", err)
				return nil, fmt.Errorf("failed to parse OpenAI response JSON: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
		}
	}

	log.Printf("🔍 Успешно распарсен ответ: статус=%s", stepResponse.Status)
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
	log.Printf("🔍 Генерируем название цели для описания: %s", description)
	log.Printf("🔍 Prompt для генерации названия цели: %s", prompt)

	response, err := c.callOpenAI(prompt)
	if err != nil {
		log.Printf("❌ Ошибка при генерации названия цели: %v", err)
		return "", fmt.Errorf("failed to call OpenAI: %w", err)
	}

	log.Printf("🔍 Получен ответ для названия цели: %s", response)

	var titleResponse struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal([]byte(response), &titleResponse); err != nil {
		log.Printf("❌ Ошибка при парсинге названия цели: %v", err)
		log.Printf("🔍 Сырой ответ: %s", response)

		// Пытаемся найти JSON в ответе
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}")

		if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
			jsonPart := response[jsonStart : jsonEnd+1]
			log.Printf("🔍 Попытка парсинга найденного JSON: %s", jsonPart)

			if err := json.Unmarshal([]byte(jsonPart), &titleResponse); err != nil {
				log.Printf("❌ Ошибка при парсинге найденного JSON: %v", err)
				return "", fmt.Errorf("failed to parse OpenAI response JSON: %w", err)
			}
		} else {
			return "", fmt.Errorf("failed to parse OpenAI response: %w", err)
		}
	}

	log.Printf("🔍 Успешно сгенерировано название: %s", titleResponse.Title)
	return titleResponse.Title, nil
}

// callOpenAI отправляет запрос к OpenAI API
func (c *OpenAIClient) callOpenAI(prompt string) (string, error) {
	// Проверяем, что API ключ установлен
	if c.apiKey == "" {
		log.Printf("❌ OpenAI API ключ не установлен")
		return "", fmt.Errorf("OpenAI API key is not set")
	}

	log.Printf("🔍 Отправляем запрос к OpenAI API...")

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
		"response_format": map[string]string{
			"type": "json_object",
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("❌ Ошибка при маршалинге запроса: %v", err)
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ Ошибка при создании HTTP запроса: %v", err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	log.Printf("🔍 Отправляем HTTP запрос к %s", c.baseURL)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("❌ Ошибка при отправке HTTP запроса: %v", err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("🔍 Получен HTTP ответ: статус %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("❌ OpenAI API вернул ошибку: %s - %s", resp.Status, string(body))
		return "", fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Ошибка при чтении тела ответа: %v", err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("🔍 Размер ответа: %d байт", len(body))

	// Парсим ответ OpenAI
	var openAIResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &openAIResponse); err != nil {
		log.Printf("❌ Ошибка при парсинге ответа OpenAI: %v", err)
		log.Printf("🔍 Сырое тело ответа: %s", string(body))
		return "", fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if len(openAIResponse.Choices) == 0 {
		log.Printf("❌ OpenAI вернул пустой список choices")
		return "", fmt.Errorf("no choices in OpenAI response")
	}

	content := openAIResponse.Choices[0].Message.Content
	log.Printf("🔍 Успешно получен контент от OpenAI: %s", content)

	return content, nil
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

	// Добавляем логику завершения цели
	prompt.WriteString("ВАЖНО: Проанализируй, достигнута ли уже цель на основе выполненных шагов.\n")
	prompt.WriteString("Если цель достигнута - верни статус 'goal_completed' и объясни почему.\n")
	prompt.WriteString("Если нужно еще 1-2 шага для завершения - верни статус 'near_completion'.\n")
	prompt.WriteString("Если цель еще далеко - верни статус 'ok' и сгенерируй следующий шаг.\n\n")

	prompt.WriteString("Сгенерируй следующий логичный шаг или определи завершение.\n\n")
	prompt.WriteString("ОТВЕТЬ СТРОГО В ФОРМАТЕ JSON:\n")
	prompt.WriteString(`{
  "status": "ok" | "need_clarification" | "goal_completed" | "near_completion",
  "step": "текст шага",
  "question": "уточняющий вопрос (если нужен)",
  "completion_reason": "причина завершения (если цель достигнута)"
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
	prompt.WriteString("ОТВЕТЬ СТРОГО В ФОРМАТЕ JSON:\n")
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
	prompt.WriteString("ОТВЕТЬ СТРОГО В ФОРМАТЕ JSON:\n")
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
	prompt.WriteString("ОТВЕТЬ СТРОГО В ФОРМАТЕ JSON:\n")
	prompt.WriteString(`{
  "title": "краткое название цели"
}`)

	return prompt.String()
}
