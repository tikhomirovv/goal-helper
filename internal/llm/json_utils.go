package llm

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// JSONParseResult представляет результат парсинга JSON
type JSONParseResult struct {
	Success bool   // Успешно ли распарсен JSON
	Content string // Извлеченный JSON контент
	Error   error  // Ошибка парсинга (если есть)
}

// ParseLLMResponse парсит ответ от LLM и извлекает JSON
// Этот метод универсален и может работать с разными LLM провайдерами
func ParseLLMResponse(response string) *JSONParseResult {
	result := &JSONParseResult{
		Success: false,
		Content: response,
	}

	// Сначала пытаемся распарсить как чистый JSON
	if isValidJSON(response) {
		result.Success = true
		result.Content = response
		return result
	}

	// Если не получилось, ищем JSON в тексте
	jsonContent, found := extractJSONFromText(response)
	if found {
		if isValidJSON(jsonContent) {
			result.Success = true
			result.Content = jsonContent
			return result
		}
	}

	// Если ничего не нашли или JSON невалидный
	result.Error = fmt.Errorf("failed to extract valid JSON from LLM response")
	return result
}

// ParseLLMResponseWithLogging парсит ответ от LLM с логированием
// Удобная обертка для ParseLLMResponse с автоматическим логированием
func ParseLLMResponseWithLogging(response string, operation string) *JSONParseResult {
	log.Printf("🔍 Парсим JSON ответ для операции: %s", operation)
	log.Printf(LogRawResponse, response)

	result := ParseLLMResponse(response)

	if !result.Success {
		log.Printf(LogParsingError, result.Error)
		return result
	}

	log.Printf("🔍 Успешно извлечен JSON: %s", result.Content)
	return result
}

// UnmarshalLLMResponse универсальный метод для парсинга JSON ответа в структуру
// Объединяет ParseLLMResponse и json.Unmarshal
func UnmarshalLLMResponse(response string, v interface{}) error {
	parseResult := ParseLLMResponse(response)
	if !parseResult.Success {
		return parseResult.Error
	}

	if err := json.Unmarshal([]byte(parseResult.Content), v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// UnmarshalLLMResponseWithLogging универсальный метод с логированием
func UnmarshalLLMResponseWithLogging(response string, v interface{}, operation string) error {
	parseResult := ParseLLMResponseWithLogging(response, operation)
	if !parseResult.Success {
		return parseResult.Error
	}

	if err := json.Unmarshal([]byte(parseResult.Content), v); err != nil {
		log.Printf(LogJSONParsingError, err)
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// isValidJSON проверяет, является ли строка валидным JSON
func isValidJSON(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

// extractJSONFromText извлекает JSON из текста
// Ищет JSON объект между фигурными скобками
func extractJSONFromText(text string) (string, bool) {
	// Ищем начало JSON объекта
	start := strings.Index(text, "{")
	if start == -1 {
		return "", false
	}

	// Ищем конец JSON объекта
	// Нужно правильно обработать вложенные скобки
	bracketCount := 0
	end := -1

	for i := start; i < len(text); i++ {
		if text[i] == '{' {
			bracketCount++
		} else if text[i] == '}' {
			bracketCount--
			if bracketCount == 0 {
				end = i
				break
			}
		}
	}

	if end == -1 || end <= start {
		return "", false
	}

	jsonContent := text[start : end+1]
	log.Printf(LogJSONParsingAttempt, jsonContent)

	return jsonContent, true
}

// ExtractJSONFromResponsesAPI извлекает JSON из ответа нового Responses API
// В новом API ответ приходит как JSON строка в content
func ExtractJSONFromResponsesAPI(response string) string {
	// В новом Responses API ответ уже является валидным JSON
	// Не нужно извлекать из обертки, как в старом API
	// Но нужно убрать возможные экранированные символы
	return response
}
