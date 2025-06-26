package llm

// JSON схемы для структурированного вывода в новом OpenAI API

// StepResponseSchema схема для ответа генерации шага
var StepResponseSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"status": map[string]any{
			"type":        "string",
			"enum":        []string{"ok", "need_clarification", "goal_completed", "near_completion"},
			"description": "Статус ответа",
		},
		"step": map[string]any{
			"type":        "string",
			"description": "Текст следующего шага",
		},
		"question": map[string]any{
			"type":        "string",
			"description": "Уточняющий вопрос (если нужен)",
		},
		"completion_reason": map[string]any{
			"type":        "string",
			"description": "Причина завершения цели (если цель достигнута)",
		},
	},
	"required":             []string{"status", "step", "question", "completion_reason"},
	"additionalProperties": false,
}

// RephraseResponseSchema схема для ответа переформулировки шага
var RephraseResponseSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"status": map[string]any{
			"type":        "string",
			"enum":        []string{"ok"},
			"description": "Статус ответа",
		},
		"step": map[string]any{
			"type":        "string",
			"description": "Новый текст шага",
		},
	},
	"required":             []string{"status", "step"},
	"additionalProperties": false,
}

// ClarificationResponseSchema схема для ответа уточнения цели
var ClarificationResponseSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"status": map[string]any{
			"type":        "string",
			"enum":        []string{"need_clarification"},
			"description": "Статус ответа",
		},
		"question": map[string]any{
			"type":        "string",
			"description": "Уточняющий вопрос",
		},
	},
	"required":             []string{"status", "question"},
	"additionalProperties": false,
}

// TitleResponseSchema схема для ответа генерации названия
var TitleResponseSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"title": map[string]any{
			"type":        "string",
			"description": "Краткое название цели",
		},
	},
	"required":             []string{"title"},
	"additionalProperties": false,
}

// ContextResponseSchema схема для ответа сбора контекста
var ContextResponseSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"status": map[string]any{
			"type":        "string",
			"enum":        []string{"ok", "need_context"},
			"description": "Статус ответа",
		},
		"question": map[string]any{
			"type":        "string",
			"description": "Конкретный вопрос для сбора контекста (если нужен)",
		},
		"context": map[string]any{
			"type":        "string",
			"description": "Краткое описание собранного контекста (если статус ok)",
		},
	},
	"required":             []string{"status", "question", "context"},
	"additionalProperties": false,
}
