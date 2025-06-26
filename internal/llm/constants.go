package llm

// Названия промптов (файлы без расширения .md)
const (
	PromptStepGeneration    = "step_generation"
	PromptStepRephrase      = "step_rephrase"
	PromptGoalClarification = "goal_clarification"
	PromptTitleGeneration   = "title_generation"
	PromptContextGathering  = "context_gathering"
)

// Статусы ответов
const (
	StatusOK                = "ok"
	StatusNeedClarification = "need_clarification"
	StatusGoalCompleted     = "goal_completed"
	StatusNearCompletion    = "near_completion"
	StatusNeedContext       = "need_context"
)

// Плейсхолдеры для промптов
const (
	PlaceholderGoalTitle       = "goal_title"
	PlaceholderGoalDescription = "goal_description"
	PlaceholderUserContext     = "user_context"
	PlaceholderCompletedSteps  = "completed_steps"
	PlaceholderCurrentStep     = "current_step"
	PlaceholderUserComment     = "user_comment"
	PlaceholderDescription     = "description"
	PlaceholderExistingContext = "existing_context"
)

// API endpoints
const (
	ResponsesAPIEndpoint   = "https://api.openai.com/v1/responses"
	CompletionsAPIEndpoint = "https://api.openai.com/v1/chat/completions"
)

// Модели по умолчанию
const (
	DefaultModel     = "gpt-4o-mini-2024-07-18"
	CompletionsModel = "gpt-4.1-nano-2025-04-14"
)

// Настройки API
const (
	DefaultTemperature = 0.7
	DefaultMaxTokens   = 500
	DefaultTimeout     = 30 // секунды
)

// Сообщения для логирования
const (
	LogSendingRequest          = "🔍 Отправляем запрос к OpenAI для цели: %s"
	LogPromptLength            = "🔍 Длина промпта: %d символов"
	LogOpenAIError             = "❌ Ошибка при вызове OpenAI: %v"
	LogOpenAIResponse          = "🔍 Получен ответ от OpenAI: %s"
	LogParsingError            = "❌ Ошибка при парсинге ответа OpenAI: %v"
	LogRawResponse             = "🔍 Сырой ответ: %s"
	LogJSONParsingAttempt      = "🔍 Попытка парсинга найденного JSON: %s"
	LogJSONParsingError        = "❌ Ошибка при парсинге найденного JSON: %v"
	LogSuccessResponse         = "🔍 Успешно распарсен ответ: статус=%s"
	LogContextGathering        = "🔍 Собираем контекст для цели: %s"
	LogContextError            = "❌ Ошибка при сборе контекста: %v"
	LogContextResponse         = "🔍 Получен ответ для сбора контекста: %s"
	LogContextSuccess          = "🔍 Успешно собран контекст: статус=%s"
	LogTitleGeneration         = "🔍 Генерируем название цели для описания: %s"
	LogTitlePrompt             = "🔍 Prompt для генерации названия цели: %s"
	LogTitleError              = "❌ Ошибка при генерации названия цели: %v"
	LogTitleResponse           = "🔍 Получен ответ для названия цели: %s"
	LogTitleParsingError       = "❌ Ошибка при парсинге названия цели: %v"
	LogTitleSuccess            = "🔍 Успешно сгенерировано название: %s"
	LogPromptLoadError         = "❌ Ошибка при загрузке промпта: %v"
	LogAPIKeyMissing           = "❌ OpenAI API ключ не установлен"
	LogSendingHTTPRequest      = "🔍 Отправляем HTTP запрос к %s"
	LogHTTPResponse            = "🔍 Получен HTTP ответ: статус %d"
	LogAPIError                = "❌ OpenAI API вернул ошибку: %s - %s"
	LogResponseSize            = "🔍 Размер ответа: %d байт"
	LogRawResponseBody         = "🔍 Сырое тело ответа: %s"
	LogNoChoices               = "❌ OpenAI вернул пустой список choices"
	LogContentReceived         = "🔍 Успешно получен контент от OpenAI: %s"
	LogMarshalingError         = "❌ Ошибка при маршалинге запроса: %v"
	LogRequestError            = "❌ Ошибка при создании HTTP запроса: %v"
	LogHTTPRequestError        = "❌ Ошибка при отправке HTTP запроса: %v"
	LogReadResponseError       = "❌ Ошибка при чтении тела ответа: %v"
	LogParseResponseError      = "❌ Ошибка при парсинге ответа OpenAI: %v"
	LogSendingRequestDetails   = "🔍 Отправляем запрос к OpenAI:"
	LogRequestURL              = "  - URL: %s"
	LogRequestModel            = "  - Model: %s"
	LogRequestBody             = "  - Request body: %s"
	LogRawAPIResponse          = "🔍 Сырой ответ от OpenAI API: %s"
	LogResponseStructure       = "🔍 Структура ответа OpenAI:"
	LogResponseID              = "  - ID: %s"
	LogResponseObject          = "  - Object: %s"
	LogResponseModel           = "  - Model: %s"
	LogResponseCreated         = "  - Created: %d"
	LogResponseUsage           = "  - Usage: %+v"
	LogResponseChoicesCount    = "  - Choices count: %d"
	LogFullResponseStructure   = "🔍 Полная структура ответа: %+v"
	LogAlternativeStructure    = "🔍 Пробуем альтернативную структуру ответа..."
	LogAltStructureError       = "❌ Ошибка при парсинге альтернативной структуры: %v"
	LogAltResponseStructure    = "🔍 Альтернативная структура ответа:"
	LogAltResponseID           = "  - ID: %s"
	LogAltResponseObject       = "  - Object: %s"
	LogAltResponseModel        = "  - Model: %s"
	LogAltResponseChoicesCount = "  - Choices count: %d"
	LogEmptyAltChoices         = "❌ Пустой список choices в альтернативной структуре"
)

// Системные сообщения для промптов
const (
	SystemMessageResponses   = "Ты помощник для достижения целей. Всегда отвечай в соответствии с указанной JSON схемой."
	SystemMessageCompletions = "Ты помощник для достижения целей. Всегда отвечай в формате JSON."
)

// Форматы для форматирования строк
const (
	FormatDescription   = "Описание: %s"
	FormatClarification = "%d. %s\n"
	FormatStep          = "%d. %s\n"
)

// JSON ключи
const (
	JSONKeyJSON = "json"
)
