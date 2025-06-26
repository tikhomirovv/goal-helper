# LLM Package

Пакет для работы с различными LLM провайдерами, включая OpenAI API.

## Архитектура

Пакет разделен на несколько компонентов:

- **OpenAI Client** (`openai.go`) - клиент для работы с OpenAI API
- **Prompt Loader** (`prompt_loader.go`) - загрузчик промптов из markdown файлов
- **Prompt Utils** (`prompt_utils.go`) - утилиты для подготовки плейсхолдеров
- **JSON Utils** (`json_utils.go`) - утилиты для парсинга JSON ответов от LLM
- **Constants** (`constants.go`) - константы для названий промптов, статусов и настроек
- **Schemas** (`schemas.go`) - JSON схемы для структурированного вывода

## Основные возможности

### 1. Поддержка двух API OpenAI

#### Новый Responses API (рекомендуется)
```go
client := NewOpenAIClientWithResponsesAPI(apiKey)
// или
client := NewOpenAIClient(apiKey) // по умолчанию использует Responses API
```

#### Старый Completions API
```go
client := NewOpenAIClientWithCompletionsAPI(apiKey)
```

### 2. Универсальные JSON утилиты

Создан универсальный механизм парсинга JSON ответов от LLM:

```go
// Простой парсинг
result := ParseLLMResponse(llmResponse)
if result.Success {
    fmt.Println("JSON:", result.Content)
}

// Парсинг с логированием
result := ParseLLMResponseWithLogging(llmResponse, "операция")

// Прямое преобразование в структуру
var response StepResponse
err := UnmarshalLLMResponse(llmResponse, &response)

// С логированием
err := UnmarshalLLMResponseWithLogging(llmResponse, &response, "генерация шага")
```

### 3. Загрузка промптов из файлов

Промпты хранятся в markdown файлах с плейсхолдерами:

```markdown
# internal/llm/prompts/step_generation.md

Ты помощник для достижения целей. Сгенерируй следующий шаг для цели.

Цель: {{{goal_title}}}
{{{goal_description}}}

{{{user_context}}}

{{{completed_steps}}}

Сгенерируй следующий шаг в формате JSON.
```

### 4. Константы для всех строковых значений

Все строковые значения вынесены в константы:

```go
// Названия промптов
PromptStepGeneration    // "step_generation"
PromptStepRephrase      // "step_rephrase"
PromptGoalClarification // "goal_clarification"

// Статусы
StatusOK                // "ok"
StatusNeedClarification // "need_clarification"

// Плейсхолдеры
PlaceholderGoalTitle    // "goal_title"
PlaceholderUserContext  // "user_context"
```

## Использование

### Создание клиента

```go
import "goal-helper/internal/llm"

// Клиент по умолчанию (Responses API)
client := llm.NewOpenAIClient(apiKey)

// С кастомной конфигурацией
config := llm.APIConfig{
    Model:   "gpt-4o-mini-2024-07-18",
    BaseURL: llm.ResponsesAPIEndpoint,
}
client := llm.NewOpenAIClientWithConfig(apiKey, config)
```

### Генерация шагов

```go
goal := &models.Goal{
    Title:       "Изучить Go",
    Description: "Изучить язык программирования Go",
}

completedSteps := []*models.Step{
    {Text: "Установить Go"},
    {Text: "Изучить синтаксис"},
}

response, err := client.GenerateStep(goal, completedSteps)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Статус: %s\n", response.Status)
fmt.Printf("Следующий шаг: %s\n", response.Step)
```

### Переформулировка шага

```go
currentStep := &models.Step{Text: "Изучить синтаксис Go"}
userComment := "Сделай более конкретным"

response, err := client.RephraseStep(goal, currentStep, userComment)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Новый шаг: %s\n", response.Step)
```

### Уточнение цели

```go
response, err := client.ClarifyGoal("Изучить Go", "Хочу изучить Go")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Вопрос: %s\n", response.Question)
```

### Генерация названия цели

```go
title, err := client.GenerateGoalTitle("Изучить язык программирования Go для веб-разработки")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Название: %s\n", title)
```

### Сбор контекста

```go
response, err := client.GatherContext(goal)
if err != nil {
    log.Fatal(err)
}

if response.Status == llm.StatusNeedContext {
    fmt.Printf("Нужен контекст: %s\n", response.Question)
} else {
    fmt.Printf("Контекст: %s\n", response.Context)
}
```

## JSON схемы

Каждый метод использует соответствующую JSON схему для структурированного вывода:

- `StepResponseSchema` - для генерации шагов
- `RephraseResponseSchema` - для переформулировки
- `ClarificationResponseSchema` - для уточнения
- `TitleResponseSchema` - для генерации названий
- `ContextResponseSchema` - для сбора контекста

## Преимущества архитектуры

✅ **Модульность** - каждый компонент отвечает за свою задачу
✅ **Переиспользуемость** - JSON утилиты можно использовать с любым LLM
✅ **Тестируемость** - каждый компонент можно тестировать отдельно
✅ **Гибкость** - легко добавить поддержку других LLM провайдеров
✅ **Чистота кода** - все константы централизованы
✅ **Структурированный вывод** - использование JSON Schema для надежности

## Тестирование

```bash
# Запуск всех тестов
go test ./internal/llm -v

# Запуск конкретного теста
go test ./internal/llm -run TestParseLLMResponse -v
```

## Структура файлов

```
internal/llm/
├── openai.go          # OpenAI клиент
├── prompt_loader.go   # Загрузчик промптов
├── prompt_utils.go    # Утилиты для плейсхолдеров
├── json_utils.go      # Утилиты для JSON
├── constants.go       # Константы
├── schemas.go         # JSON схемы
├── types.go           # Типы данных
├── README.md          # Документация
└── prompts/           # Промпты в markdown
    ├── step_generation.md
    ├── step_rephrase.md
    ├── goal_clarification.md
    ├── title_generation.md
    └── context_gathering.md
```
