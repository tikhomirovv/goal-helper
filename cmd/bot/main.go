package main

import (
	"log"
	"os"

	"goal-helper/internal/bot"
	"goal-helper/internal/llm"
	"goal-helper/internal/repository"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные окружения из .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Получаем токен бота из переменных окружения
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN environment variable is required")
	}

	// Инициализируем репозиторий для работы с данными
	repo, err := repository.NewFileRepository("data")
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	// Инициализируем LLM клиент
	llmClient := llm.NewClient(os.Getenv("LLM_API_KEY"))

	// Создаем и запускаем бота
	botInstance := bot.NewBot(botToken, repo, llmClient)

	log.Println("Starting Goal Helper bot...")
	if err := botInstance.Start(); err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
}
