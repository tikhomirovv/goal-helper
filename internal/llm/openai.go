package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"goal-helper/internal/models"
)

// OpenAIClient представляет клиент для работы с OpenAI API
type OpenAIClient struct {
	apiKey       string
	httpClient   *http.Client
	baseURL      string
	model        string
	promptLoader *PromptLoader // Загрузчик промптов из файлов
	promptUtils  *PromptUtils  // Утилиты для подготовки плейсхолдеров
}

// APIConfig представляет конфигурацию для API запроса
type APIConfig struct {
	Model   string // Модель для использования
	BaseURL string // Базовый URL (responses или completions)
}

// DefaultAPIConfig возвращает конфигурацию по умолчанию
func DefaultAPIConfig() APIConfig {
	return APIConfig{
		Model:   DefaultModel,
		BaseURL: ResponsesAPIEndpoint, // Новый endpoint для структурированного вывода
	}
}

// CompletionsAPIConfig возвращает конфигурацию для старого completions API
func CompletionsAPIConfig() APIConfig {
	return APIConfig{
		Model:   CompletionsModel,
		BaseURL: CompletionsAPIEndpoint,
	}
}

// NewOpenAIClient создает новый OpenAI клиент
func NewOpenAIClient(apiKey string) Client {
	return NewOpenAIClientWithResponsesAPI(apiKey)
}

// NewOpenAIClientWithConfig создает новый OpenAI клиент с кастомной конфигурацией
func NewOpenAIClientWithConfig(apiKey string, config APIConfig) Client {
	return &OpenAIClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: DefaultTimeout * time.Second,
		},
		baseURL:      config.BaseURL,
		model:        config.Model,
		promptLoader: NewPromptLoader(),
		promptUtils:  NewPromptUtils(),
	}
}

// NewOpenAIClientWithCompletionsAPI создает клиент для работы со старым Completions API
func NewOpenAIClientWithCompletionsAPI(apiKey string) Client {
	return NewOpenAIClientWithConfig(apiKey, CompletionsAPIConfig())
}

// NewOpenAIClientWithResponsesAPI создает клиент для работы с новым Responses API
func NewOpenAIClientWithResponsesAPI(apiKey string) Client {
	return NewOpenAIClientWithConfig(apiKey, DefaultAPIConfig())
}

// GenerateStepWithConfig генерирует следующий шаг для цели с кастомной конфигурацией
func (c *OpenAIClient) GenerateStepWithConfig(goal *models.Goal, completedSteps []*models.Step, config APIConfig) (*StepResponse, error) {
	// Загружаем промпт из файла
	placeholders := c.promptUtils.BuildStepPromptPlaceholders(goal, completedSteps)
	prompt, err := c.promptLoader.LoadPrompt(PromptStepGeneration, placeholders)
	if err != nil {
		log.Printf(LogPromptLoadError, err)
		return nil, fmt.Errorf("failed to load prompt: %w", err)
	}

	log.Printf(LogSendingRequest, goal.Title)
	log.Printf(LogPromptLength, len(prompt))

	response, err := c.callOpenAI(prompt, config, StepResponseSchema)
	if err != nil {
		log.Printf(LogOpenAIError, err)
		return nil, fmt.Errorf("failed to call OpenAI: %w", err)
	}

	log.Printf(LogOpenAIResponse, response)

	var stepResponse StepResponse
	if err := UnmarshalLLMResponseWithLogging(response, &stepResponse, "генерация шага"); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	log.Printf(LogSuccessResponse, stepResponse.Status)
	return &stepResponse, nil
}

// GenerateStep генерирует следующий шаг для цели
func (c *OpenAIClient) GenerateStep(goal *models.Goal, completedSteps []*models.Step) (*StepResponse, error) {
	return c.GenerateStepWithConfig(goal, completedSteps, DefaultAPIConfig())
}

// RephraseStep переформулирует текущий шаг
func (c *OpenAIClient) RephraseStep(goal *models.Goal, currentStep *models.Step, userComment string) (*StepResponse, error) {
	// Загружаем промпт из файла
	placeholders := c.promptUtils.BuildRephrasePromptPlaceholders(goal, currentStep, userComment)

	prompt, err := c.promptLoader.LoadPrompt(PromptStepRephrase, placeholders)
	if err != nil {
		log.Printf(LogPromptLoadError, err)
		return nil, fmt.Errorf("failed to load prompt: %w", err)
	}

	response, err := c.callOpenAI(prompt, DefaultAPIConfig(), RephraseResponseSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI: %w", err)
	}

	var stepResponse StepResponse
	if err := UnmarshalLLMResponseWithLogging(response, &stepResponse, "переформулировка шага"); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	return &stepResponse, nil
}

// ClarifyGoal запрашивает уточнение цели
func (c *OpenAIClient) ClarifyGoal(goalTitle, goalDescription string) (*ClarificationResponse, error) {
	// Загружаем промпт из файла
	placeholders := c.promptUtils.BuildClarificationPromptPlaceholders(goalTitle, goalDescription)

	prompt, err := c.promptLoader.LoadPrompt(PromptGoalClarification, placeholders)
	if err != nil {
		log.Printf(LogPromptLoadError, err)
		return nil, fmt.Errorf("failed to load prompt: %w", err)
	}

	response, err := c.callOpenAI(prompt, DefaultAPIConfig(), ClarificationResponseSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI: %w", err)
	}

	var clarificationResponse ClarificationResponse
	if err := UnmarshalLLMResponseWithLogging(response, &clarificationResponse, "уточнение цели"); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	return &clarificationResponse, nil
}

// GenerateGoalTitle генерирует название цели на основе описания
func (c *OpenAIClient) GenerateGoalTitle(description string) (string, error) {
	// Загружаем промпт из файла
	placeholders := c.promptUtils.BuildTitlePromptPlaceholders(description)

	prompt, err := c.promptLoader.LoadPrompt(PromptTitleGeneration, placeholders)
	if err != nil {
		log.Printf(LogPromptLoadError, err)
		return "", fmt.Errorf("failed to load prompt: %w", err)
	}

	log.Printf(LogTitleGeneration, description)
	log.Printf(LogTitlePrompt, prompt)

	response, err := c.callOpenAI(prompt, DefaultAPIConfig(), TitleResponseSchema)
	if err != nil {
		log.Printf(LogTitleError, err)
		return "", fmt.Errorf("failed to call OpenAI: %w", err)
	}

	log.Printf(LogTitleResponse, response)

	var titleResponse struct {
		Title string `json:"title"`
	}
	if err := UnmarshalLLMResponseWithLogging(response, &titleResponse, "генерация названия цели"); err != nil {
		return "", fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	log.Printf(LogTitleSuccess, titleResponse.Title)
	return titleResponse.Title, nil
}

// GatherContext собирает контекст пользователя для более точной генерации шагов
func (c *OpenAIClient) GatherContext(goal *models.Goal) (*ContextResponse, error) {
	// Загружаем промпт из файла
	placeholders := c.promptUtils.BuildContextPromptPlaceholders(goal)
	prompt, err := c.promptLoader.LoadPrompt(PromptContextGathering, placeholders)
	if err != nil {
		log.Printf(LogPromptLoadError, err)
		return nil, fmt.Errorf("failed to load prompt: %w", err)
	}

	log.Printf(LogContextGathering, goal.Title)

	response, err := c.callOpenAI(prompt, DefaultAPIConfig(), ContextResponseSchema)
	if err != nil {
		log.Printf(LogContextError, err)
		return nil, fmt.Errorf("failed to call OpenAI: %w", err)
	}

	log.Printf(LogContextResponse, response)

	var contextResponse ContextResponse
	if err := UnmarshalLLMResponseWithLogging(response, &contextResponse, "сбор контекста"); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	log.Printf(LogContextSuccess, contextResponse.Status)
	return &contextResponse, nil
}

// callOpenAI отправляет запрос к OpenAI API с поддержкой нового Responses API
func (c *OpenAIClient) callOpenAI(prompt string, config APIConfig, responseSchema map[string]any) (string, error) {
	// Проверяем, что API ключ установлен
	if c.apiKey == "" {
		log.Printf(LogAPIKeyMissing)
		return "", fmt.Errorf("OpenAI API key is not set")
	}

	log.Printf("🔍 Отправляем запрос к OpenAI API...")

	// Определяем, какой API использовать
	var requestBody map[string]any

	if config.BaseURL == ResponsesAPIEndpoint {
		// Новый Responses API с JSON Schema
		requestBody = map[string]any{
			"model": config.Model,
			"input": []map[string]string{
				{
					"role":    "system",
					"content": SystemMessageResponses,
				},
				{
					"role":    "user",
					"content": prompt,
				},
			},
			"text": map[string]any{
				"format": map[string]any{
					"type":   "json_schema",
					"name":   "goal_assistant_response",
					"schema": responseSchema,
					"strict": true,
				},
			},
		}
	} else {
		// Старый Completions API
		requestBody = map[string]any{
			"model": config.Model,
			"messages": []map[string]string{
				{
					"role":    "system",
					"content": SystemMessageCompletions,
				},
				{
					"role":    "user",
					"content": prompt,
				},
			},
			"temperature": DefaultTemperature,
			"max_tokens":  DefaultMaxTokens,
			"response_format": map[string]string{
				"type": "json_object",
			},
		}
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf(LogMarshalingError, err)
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Логируем запрос для диагностики
	log.Printf(LogSendingRequestDetails)
	log.Printf(LogRequestURL, config.BaseURL)
	log.Printf(LogRequestModel, config.Model)
	log.Printf(LogRequestBody, string(jsonData))

	req, err := http.NewRequest("POST", config.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf(LogRequestError, err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	log.Printf(LogSendingHTTPRequest, config.BaseURL)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf(LogHTTPRequestError, err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf(LogHTTPResponse, resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf(LogAPIError, resp.Status, string(body))
		return "", fmt.Errorf("OpenAI API error: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf(LogReadResponseError, err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf(LogResponseSize, len(body))

	// Логируем сырой ответ для диагностики
	log.Printf(LogRawAPIResponse, string(body))

	// Парсим ответ OpenAI
	if config.BaseURL == ResponsesAPIEndpoint {
		// Новый Responses API имеет другую структуру
		var responsesAPIResponse struct {
			ID     string `json:"id"`
			Object string `json:"object"`
			Status string `json:"status"`
			Model  string `json:"model"`
			Output []struct {
				ID      string `json:"id"`
				Type    string `json:"type"`
				Status  string `json:"status"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
				Role string `json:"role"`
			} `json:"output"`
			Usage struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
				TotalTokens  int `json:"total_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal(body, &responsesAPIResponse); err != nil {
			log.Printf(LogParseResponseError, err)
			log.Printf(LogRawResponseBody, string(body))
			return "", fmt.Errorf("failed to parse OpenAI Responses API response: %w", err)
		}

		// Логируем структуру ответа для диагностики
		log.Printf(LogResponseStructure)
		log.Printf(LogResponseID, responsesAPIResponse.ID)
		log.Printf(LogResponseObject, responsesAPIResponse.Object)
		log.Printf(LogResponseModel, responsesAPIResponse.Model)
		log.Printf("  - Status: %s", responsesAPIResponse.Status)
		log.Printf("  - Output count: %d", len(responsesAPIResponse.Output))
		log.Printf(LogResponseUsage, responsesAPIResponse.Usage)

		if len(responsesAPIResponse.Output) == 0 {
			log.Printf("❌ OpenAI Responses API вернул пустой список output")
			return "", fmt.Errorf("no output in OpenAI Responses API response")
		}

		// Извлекаем контент из первого output
		output := responsesAPIResponse.Output[0]
		if len(output.Content) == 0 {
			log.Printf("❌ Пустой контент в output")
			return "", fmt.Errorf("empty content in OpenAI Responses API output")
		}

		// Ищем текстовый контент
		var content string
		for _, contentItem := range output.Content {
			if contentItem.Type == "output_text" {
				content = contentItem.Text
				break
			}
		}

		if content == "" {
			log.Printf("❌ Не найден текстовый контент в output")
			return "", fmt.Errorf("no text content found in OpenAI Responses API output")
		}

		log.Printf(LogContentReceived, content)
		return content, nil

	} else {
		// Старый Completions API
		var openAIResponse struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			Model   string `json:"model"`
			Usage   struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal(body, &openAIResponse); err != nil {
			log.Printf(LogParseResponseError, err)
			log.Printf(LogRawResponseBody, string(body))
			return "", fmt.Errorf("failed to parse OpenAI response: %w", err)
		}

		// Логируем структуру ответа для диагностики
		log.Printf(LogResponseStructure)
		log.Printf(LogResponseID, openAIResponse.ID)
		log.Printf(LogResponseObject, openAIResponse.Object)
		log.Printf(LogResponseModel, openAIResponse.Model)
		log.Printf(LogResponseCreated, openAIResponse.Created)
		log.Printf(LogResponseUsage, openAIResponse.Usage)
		log.Printf(LogResponseChoicesCount, len(openAIResponse.Choices))

		if len(openAIResponse.Choices) == 0 {
			log.Printf(LogNoChoices)
			log.Printf(LogFullResponseStructure, openAIResponse)
			return "", fmt.Errorf("no choices in OpenAI response")
		}

		content := openAIResponse.Choices[0].Message.Content
		log.Printf(LogContentReceived, content)

		// Для старого API извлекаем JSON из структуры ответа
		content = ExtractJSONFromResponsesAPI(content)
		return content, nil
	}
}
