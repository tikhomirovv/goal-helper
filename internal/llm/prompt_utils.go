package llm

import (
	"fmt"
	"strings"

	"goal-helper/internal/models"
)

// PromptUtils предоставляет утилиты для подготовки плейсхолдеров промптов
type PromptUtils struct{}

// NewPromptUtils создает новый экземпляр утилит для промптов
func NewPromptUtils() *PromptUtils {
	return &PromptUtils{}
}

// BuildStepPromptPlaceholders подготавливает плейсхолдеры для промпта генерации шагов
func (pu *PromptUtils) BuildStepPromptPlaceholders(goal *models.Goal, completedSteps []*models.Step) map[string]string {
	placeholders := make(map[string]string)

	// Основная информация о цели
	placeholders[PlaceholderGoalTitle] = goal.Title
	if goal.Description != "" {
		placeholders[PlaceholderGoalDescription] = fmt.Sprintf(FormatDescription, goal.Description)
	} else {
		placeholders[PlaceholderGoalDescription] = ""
	}

	// Контекст пользователя
	if len(goal.Context.Clarifications) > 0 {
		var contextBuilder strings.Builder
		for i, clarification := range goal.Context.Clarifications {
			contextBuilder.WriteString(fmt.Sprintf(FormatClarification, i+1, clarification))
		}
		placeholders[PlaceholderUserContext] = contextBuilder.String()
	} else {
		placeholders[PlaceholderUserContext] = ""
	}

	// Выполненные шаги
	if len(completedSteps) > 0 {
		var stepsBuilder strings.Builder
		for i, step := range completedSteps {
			stepsBuilder.WriteString(fmt.Sprintf(FormatStep, i+1, step.Text))
		}
		placeholders[PlaceholderCompletedSteps] = stepsBuilder.String()
	} else {
		placeholders[PlaceholderCompletedSteps] = ""
	}

	return placeholders
}

// BuildContextPromptPlaceholders подготавливает плейсхолдеры для промпта сбора контекста
func (pu *PromptUtils) BuildContextPromptPlaceholders(goal *models.Goal) map[string]string {
	placeholders := make(map[string]string)

	// Основная информация о цели
	placeholders[PlaceholderGoalTitle] = goal.Title
	if goal.Description != "" {
		placeholders[PlaceholderGoalDescription] = fmt.Sprintf(FormatDescription, goal.Description)
	} else {
		placeholders[PlaceholderGoalDescription] = ""
	}

	// Уже собранный контекст
	if len(goal.Context.Clarifications) > 0 {
		var contextBuilder strings.Builder
		for i, clarification := range goal.Context.Clarifications {
			contextBuilder.WriteString(fmt.Sprintf(FormatClarification, i+1, clarification))
		}
		placeholders[PlaceholderExistingContext] = contextBuilder.String()
	} else {
		placeholders[PlaceholderExistingContext] = ""
	}

	return placeholders
}

// BuildRephrasePromptPlaceholders подготавливает плейсхолдеры для промпта переформулировки
func (pu *PromptUtils) BuildRephrasePromptPlaceholders(goal *models.Goal, currentStep *models.Step, userComment string) map[string]string {
	return map[string]string{
		PlaceholderGoalTitle:   goal.Title,
		PlaceholderCurrentStep: currentStep.Text,
		PlaceholderUserComment: userComment,
	}
}

// BuildClarificationPromptPlaceholders подготавливает плейсхолдеры для промпта уточнения
func (pu *PromptUtils) BuildClarificationPromptPlaceholders(goalTitle, goalDescription string) map[string]string {
	placeholders := map[string]string{
		PlaceholderGoalTitle: goalTitle,
	}

	if goalDescription != "" {
		placeholders[PlaceholderGoalDescription] = fmt.Sprintf(FormatDescription, goalDescription)
	} else {
		placeholders[PlaceholderGoalDescription] = ""
	}

	return placeholders
}

// BuildTitlePromptPlaceholders подготавливает плейсхолдеры для промпта генерации названия
func (pu *PromptUtils) BuildTitlePromptPlaceholders(description string) map[string]string {
	return map[string]string{
		PlaceholderDescription: description,
	}
}
