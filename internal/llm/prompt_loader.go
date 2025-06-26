package llm

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PromptLoader представляет загрузчик промптов из файлов
type PromptLoader struct {
	promptsDir string
	cache      map[string]string // Кэш для загруженных промптов
}

// NewPromptLoader создает новый загрузчик промптов
func NewPromptLoader() *PromptLoader {
	// Определяем путь к папке с промптами относительно текущего файла
	currentDir, _ := os.Getwd()
	promptsDir := filepath.Join(currentDir, "internal", "llm", "prompts")

	// Если не найдено, пробуем относительный путь
	if _, err := os.Stat(promptsDir); os.IsNotExist(err) {
		promptsDir = "internal/llm/prompts"
	}

	return &PromptLoader{
		promptsDir: promptsDir,
		cache:      make(map[string]string),
	}
}

// LoadPrompt загружает промпт из файла и подставляет значения
// filename - имя файла без расширения (например, "step_generation")
// placeholders - карта плейсхолдеров для подстановки
func (pl *PromptLoader) LoadPrompt(filename string, placeholders map[string]string) (string, error) {
	// Проверяем кэш
	if cached, exists := pl.cache[filename]; exists {
		return pl.replacePlaceholders(cached, placeholders), nil
	}

	// Формируем полный путь к файлу
	filePath := filepath.Join(pl.promptsDir, filename+".md")

	// Читаем файл
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file %s: %w", filePath, err)
	}

	// Убираем заголовок markdown (первую строку с #)
	lines := strings.Split(string(content), "\n")
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "#") {
		lines = lines[1:]
	}

	// Объединяем строки обратно
	promptContent := strings.TrimSpace(strings.Join(lines, "\n"))

	// Кэшируем промпт
	pl.cache[filename] = promptContent

	// Подставляем плейсхолдеры и возвращаем результат
	return pl.replacePlaceholders(promptContent, placeholders), nil
}

// replacePlaceholders заменяет плейсхолдеры в формате {{{key}}} на соответствующие значения
func (pl *PromptLoader) replacePlaceholders(content string, placeholders map[string]string) string {
	result := content

	// Регулярное выражение для поиска плейсхолдеров в формате {{{key}}}
	placeholderRegex := regexp.MustCompile(`\{\{\{(\w+)\}\}\}`)

	// Заменяем все найденные плейсхолдеры
	result = placeholderRegex.ReplaceAllStringFunc(result, func(match string) string {
		// Извлекаем ключ из плейсхолдера (убираем {{{ и }}})
		key := match[3 : len(match)-3]

		// Если значение найдено, возвращаем его, иначе оставляем плейсхолдер как есть
		if value, exists := placeholders[key]; exists {
			return value
		}

		// Если плейсхолдер не найден, возвращаем пустую строку
		return ""
	})

	return result
}

// ListAvailablePrompts возвращает список доступных промптов
func (pl *PromptLoader) ListAvailablePrompts() ([]string, error) {
	var prompts []string

	err := filepath.WalkDir(pl.promptsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			// Убираем расширение .md
			promptName := strings.TrimSuffix(d.Name(), ".md")
			prompts = append(prompts, promptName)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list prompts: %w", err)
	}

	return prompts, nil
}

// ClearCache очищает кэш промптов
func (pl *PromptLoader) ClearCache() {
	pl.cache = make(map[string]string)
}
