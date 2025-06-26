package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"goal-helper/internal/llm"
	"goal-helper/internal/models"
	"goal-helper/internal/repository"

	tele "gopkg.in/telebot.v3"
)

// Bot представляет Telegram-бота
type Bot struct {
	bot       *tele.Bot
	repo      repository.Repository
	llmClient llm.Client
	states    map[int64]*UserState // Состояния пользователей
}

// UserState представляет состояние пользователя в FSM
type UserState struct {
	UserID   int64
	State    string            // "idle", "waiting_goal_description", "rephrasing"
	TempData map[string]string // Временные данные для создания цели
}

// NewBot создает нового бота
func NewBot(token string, repo repository.Repository, llmClient llm.Client) *Bot {
	// Настройки бота
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	b := &Bot{
		bot:       bot,
		repo:      repo,
		llmClient: llmClient,
		states:    make(map[int64]*UserState),
	}

	// Регистрируем обработчики команд
	b.registerHandlers()

	return b
}

// Start запускает бота
func (b *Bot) Start() error {
	log.Println("Bot started...")
	b.bot.Start()
	return nil
}

// registerHandlers регистрирует все обработчики команд
func (b *Bot) registerHandlers() {
	// Основные команды
	b.bot.Handle("/start", b.handleStart)
	b.bot.Handle("/help", b.handleHelp)
	b.bot.Handle("/goals", b.handleGoals)
	b.bot.Handle("/newgoal", b.handleNewGoal)
	b.bot.Handle("/status", b.handleStatus)
	b.bot.Handle("/step", b.handleStep)
	b.bot.Handle("/done", b.handleDone)
	b.bot.Handle("/next", b.handleNext)
	b.bot.Handle("/rephrase", b.handleRephrase)
	b.bot.Handle("/switch", b.handleSwitch)
	b.bot.Handle("/complete", b.handleComplete)

	// Обработчик кнопок
	b.bot.Handle(&tele.Btn{Text: "✅ Выполнил"}, b.handleDone)
	b.bot.Handle(&tele.Btn{Text: "🔄 Переформулировать"}, b.handleRephrase)
	b.bot.Handle(&tele.Btn{Text: "🎉 Завершить цель"}, b.handleComplete)

	// Обработка текстовых сообщений
	b.bot.Handle(tele.OnText, b.handleText)
}

// handleStart обрабатывает команду /start
func (b *Bot) handleStart(c tele.Context) error {
	userID := strconv.FormatInt(c.Sender().ID, 10)
	username := c.Sender().Username
	firstName := c.Sender().FirstName

	// Проверяем, существует ли пользователь
	_, err := b.repo.GetUser(userID)
	if err != nil {
		// Создаем нового пользователя
		user := models.NewUser(userID, username, firstName)
		if err := b.repo.CreateUser(user); err != nil {
			return c.Send("❌ Ошибка при создании пользователя")
		}
	}

	// Создаем или обновляем состояние пользователя
	b.states[c.Sender().ID] = &UserState{
		UserID:   c.Sender().ID,
		State:    "idle",
		TempData: make(map[string]string),
	}

	// Приветственное сообщение
	message := fmt.Sprintf("🎯 Привет, %s!\n\nЯ помогу тебе достичь целей через простые шаги.\n\nЧто хочешь сделать?", firstName)

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	btnGoals := menu.Text("📋 Мои цели")
	btnNewGoal := menu.Text("➕ Новая цель")

	menu.Reply(
		menu.Row(btnGoals),
		menu.Row(btnNewGoal),
	)

	return c.Send(message, menu)
}

// handleHelp обрабатывает команду /help
func (b *Bot) handleHelp(c tele.Context) error {
	message := `🤖 **Помощник в достижении целей**

**Основные команды:**
/start - Начать работу с ботом
/help - Показать эту справку
/goals - Показать список твоих целей
/newgoal - Создать новую цель
/status - Показать прогресс по активной цели
/step - Показать текущий шаг
/done - Отметить шаг как выполненный
/next - Получить следующий шаг
/rephrase - Переформулировать текущий шаг
/complete - Завершить цель (если считаешь, что она достигнута)
/switch - Переключиться на другую цель

**Как это работает:**
1. Создай цель командой /newgoal
2. Получи первый шаг командой /next
3. Выполни шаг и отметь его командой /done
4. Получи следующий шаг командой /next
5. Повторяй, пока цель не будет достигнута

**Статусы целей:**
🎯 - Активная цель
✅ - Завершенная цель
⏳ - Неактивная цель

Бот сам определит, когда цель достигнута, но ты можешь завершить её вручную командой /complete.`

	return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

// handleGoals обрабатывает команду /goals
func (b *Bot) handleGoals(c tele.Context) error {
	userID := strconv.FormatInt(c.Sender().ID, 10)

	goals, err := b.repo.GetUserGoals(userID)
	if err != nil {
		return c.Send("❌ Ошибка при получении целей")
	}

	if len(goals) == 0 {
		return c.Send("📝 У тебя пока нет целей.\n\nСоздай первую цель командой /newgoal")
	}

	var message strings.Builder
	message.WriteString("📋 **Твои цели:**\n\n")

	user, err := b.repo.GetUser(userID)
	if err != nil {
		return c.Send("❌ Ошибка при получении данных пользователя")
	}

	for i, goal := range goals {
		status := "⏳"
		if goal.Status == "completed" {
			status = "✅"
		} else if goal.ID == user.ActiveGoalID {
			status = "🎯"
		}
		message.WriteString(fmt.Sprintf("%s **%d. %s**\n", status, i+1, goal.Title))
		if goal.Description != "" {
			message.WriteString(fmt.Sprintf("   %s\n", goal.Description))
		}
		message.WriteString("\n")
	}

	return c.Send(message.String(), &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

// handleNewGoal обрабатывает команду /newgoal
func (b *Bot) handleNewGoal(c tele.Context) error {
	// Устанавливаем состояние "создание цели"
	state := b.getOrCreateState(c.Sender().ID)
	state.State = "waiting_goal_description"
	state.TempData = make(map[string]string)

	return c.Send("🎯 Отлично! Давай создадим новую цель.\n\nОпиши свою цель подробно - что именно ты хочешь достичь? Я сам придумаю подходящее название.")
}

// handleStatus обрабатывает команду /status
func (b *Bot) handleStatus(c tele.Context) error {
	userID := strconv.FormatInt(c.Sender().ID, 10)

	user, err := b.repo.GetUser(userID)
	if err != nil {
		return c.Send("❌ Ошибка при получении данных пользователя")
	}

	if user.ActiveGoalID == "" {
		return c.Send("📝 У тебя нет активной цели.\n\nВыбери цель из списка командой /goals или создай новую командой /newgoal")
	}

	goal, err := b.repo.GetGoal(user.ActiveGoalID)
	if err != nil {
		return c.Send("❌ Ошибка при получении активной цели")
	}

	// Проверяем статус цели
	if goal.Status == "completed" {
		return c.Send("✅ Эта цель уже завершена!\n\nСоздай новую цель командой /newgoal или выбери другую из списка /goals")
	}

	steps, err := b.repo.GetGoalSteps(goal.ID)
	if err != nil {
		return c.Send("❌ Ошибка при получении шагов")
	}

	completedCount := 0
	for _, step := range steps {
		if step.IsCompleted() {
			completedCount++
		}
	}

	message := fmt.Sprintf("🎯 **Активная цель:** %s\n\n", goal.Title)
	if goal.Description != "" {
		message += fmt.Sprintf("📝 %s\n\n", goal.Description)
	}
	message += fmt.Sprintf("📊 **Прогресс:** %d/%d шагов выполнено\n\n", completedCount, len(steps))
	message += "Используй /step чтобы увидеть текущий шаг"

	return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

// handleStep обрабатывает команду /step
func (b *Bot) handleStep(c tele.Context) error {
	userID := strconv.FormatInt(c.Sender().ID, 10)

	user, err := b.repo.GetUser(userID)
	if err != nil {
		return c.Send("❌ Ошибка при получении данных пользователя")
	}

	if user.ActiveGoalID == "" {
		return c.Send("📝 У тебя нет активной цели.\n\nВыбери цель из списка командой /goals")
	}

	goal, err := b.repo.GetGoal(user.ActiveGoalID)
	if err != nil {
		return c.Send("❌ Ошибка при получении активной цели")
	}

	// Проверяем статус цели
	if goal.Status == "completed" {
		return c.Send("✅ Эта цель уже завершена!\n\nСоздай новую цель командой /newgoal или выбери другую из списка /goals")
	}

	currentStep, err := b.repo.GetCurrentStep(user.ActiveGoalID)
	if err != nil {
		return c.Send("✅ Поздравляю! Ты выполнил все шаги для этой цели.\n\nИспользуй /next чтобы получить следующий шаг")
	}

	message := fmt.Sprintf("📝 **Текущий шаг:**\n\n%s", currentStep.Text)

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	btnDone := menu.Text("✅ Выполнил")
	btnRephrase := menu.Text("🔄 Переформулировать")

	menu.Reply(
		menu.Row(btnDone),
		menu.Row(btnRephrase),
	)

	return c.Send(message, menu, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

// handleDone обрабатывает команду /done
func (b *Bot) handleDone(c tele.Context) error {
	userID := strconv.FormatInt(c.Sender().ID, 10)

	user, err := b.repo.GetUser(userID)
	if err != nil {
		return c.Send("❌ Ошибка при получении данных пользователя")
	}

	if user.ActiveGoalID == "" {
		return c.Send("📝 У тебя нет активной цели")
	}

	goal, err := b.repo.GetGoal(user.ActiveGoalID)
	if err != nil {
		return c.Send("❌ Ошибка при получении цели")
	}

	// Проверяем статус цели
	if goal.Status == "completed" {
		return c.Send("✅ Эта цель уже завершена!\n\nСоздай новую цель командой /newgoal или выбери другую из списка /goals")
	}

	currentStep, err := b.repo.GetCurrentStep(user.ActiveGoalID)
	if err != nil {
		return c.Send("✅ Поздравляю! Ты выполнил все шаги для этой цели")
	}

	// Отмечаем шаг как выполненный
	currentStep.Complete()
	if err := b.repo.UpdateStep(currentStep); err != nil {
		return c.Send("❌ Ошибка при обновлении шага")
	}

	return c.Send("✅ Отлично! Шаг выполнен.\n\nИспользуй /next чтобы получить следующий шаг")
}

// handleNext обрабатывает команду /next
func (b *Bot) handleNext(c tele.Context) error {
	userID := strconv.FormatInt(c.Sender().ID, 10)

	user, err := b.repo.GetUser(userID)
	if err != nil {
		return c.Send("❌ Ошибка при получении данных пользователя")
	}

	if user.ActiveGoalID == "" {
		return c.Send("📝 У тебя нет активной цели")
	}

	goal, err := b.repo.GetGoal(user.ActiveGoalID)
	if err != nil {
		return c.Send("❌ Ошибка при получении цели")
	}

	// Проверяем статус цели
	if goal.Status == "completed" {
		return c.Send("✅ Эта цель уже завершена!\n\nСоздай новую цель командой /newgoal или выбери другую из списка /goals")
	}

	// Получаем все шаги для цели
	allSteps, err := b.repo.GetGoalSteps(goal.ID)
	if err != nil {
		return c.Send("❌ Ошибка при получении шагов")
	}

	// Проверяем, есть ли невыполненные шаги
	var currentStep *models.Step
	var completedSteps []*models.Step

	for _, step := range allSteps {
		if step.IsCompleted() {
			completedSteps = append(completedSteps, step)
		} else {
			// Нашли невыполненный шаг
			if currentStep == nil || step.CreatedAt.Before(currentStep.CreatedAt) {
				currentStep = step
			}
		}
	}

	// Если есть невыполненный шаг, предлагаем его выполнить
	if currentStep != nil {
		message := fmt.Sprintf("⏳ У тебя есть невыполненный шаг:\n\n**%s**\n\nСначала выполни этот шаг командой /done, а потом получи следующий.", currentStep.Text)
		return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	// Все шаги выполнены, генерируем следующий
	log.Printf("🔍 Генерируем следующий шаг для цели: %s", goal.Title)
	log.Printf("🔍 Количество выполненных шагов: %d", len(completedSteps))

	response, err := b.llmClient.GenerateStep(goal, completedSteps)
	if err != nil {
		log.Printf("❌ Ошибка при генерации шага: %v", err)
		return c.Send(fmt.Sprintf("❌ Ошибка при генерации шага: %v", err))
	}

	log.Printf("🔍 Получен ответ от LLM: статус=%s, шаг=%s", response.Status, response.Step)

	if response.Status == "need_clarification" {
		return c.Send(fmt.Sprintf("❓ %s", response.Question))
	}

	// Обрабатываем завершение цели
	if response.Status == "goal_completed" {
		if err := b.completeGoal(goal, user, response.CompletionReason); err != nil {
			return c.Send("❌ Ошибка при завершении цели")
		}

		message := fmt.Sprintf("🎉 **Поздравляю! Цель достигнута!**\n\n**%s**\n\n%s\n\nСоздай новую цель командой /newgoal",
			goal.Title, response.CompletionReason)
		return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	// Обрабатываем близость к завершению
	if response.Status == "near_completion" {
		// Создаем новый шаг
		newStep := models.NewStep(goal.ID, response.Step)
		if err := b.repo.CreateStep(newStep); err != nil {
			return c.Send("❌ Ошибка при создании шага")
		}

		message := fmt.Sprintf("🎯 **Почти готово! Осталось совсем немного:**\n\n%s\n\n💡 После этого шага цель может быть достигнута!", newStep.Text)

		menu := &tele.ReplyMarkup{ResizeKeyboard: true}
		btnDone := menu.Text("✅ Выполнил")
		btnRephrase := menu.Text("🔄 Переформулировать")
		btnComplete := menu.Text("🎉 Завершить цель")

		menu.Reply(
			menu.Row(btnDone),
			menu.Row(btnRephrase),
			menu.Row(btnComplete),
		)

		return c.Send(message, menu, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	// Обычный шаг
	if response.Status == "ok" {
		// Создаем новый шаг
		newStep := models.NewStep(goal.ID, response.Step)
		if err := b.repo.CreateStep(newStep); err != nil {
			return c.Send("❌ Ошибка при создании шага")
		}

		message := fmt.Sprintf("📝 **Новый шаг:**\n\n%s", newStep.Text)

		menu := &tele.ReplyMarkup{ResizeKeyboard: true}
		btnDone := menu.Text("✅ Выполнил")
		btnRephrase := menu.Text("🔄 Переформулировать")

		menu.Reply(
			menu.Row(btnDone),
			menu.Row(btnRephrase),
		)

		return c.Send(message, menu, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	// Неизвестный статус
	return c.Send("❌ Неожиданный ответ от системы")
}

// handleRephrase обрабатывает команду /rephrase
func (b *Bot) handleRephrase(c tele.Context) error {
	// Устанавливаем состояние "переформулировка"
	state := b.getOrCreateState(c.Sender().ID)
	state.State = "rephrasing"
	state.TempData = make(map[string]string)

	return c.Send("🔄 Опиши, что именно не подходит в текущем шаге?\n\nНапример: \"Слишком сложно\", \"Непонятно что делать\", \"Нужно что-то проще\"")
}

// handleSwitch обрабатывает команду /switch
func (b *Bot) handleSwitch(c tele.Context) error {
	userID := strconv.FormatInt(c.Sender().ID, 10)

	goals, err := b.repo.GetUserGoals(userID)
	if err != nil {
		return c.Send("❌ Ошибка при получении целей")
	}

	if len(goals) == 0 {
		return c.Send("📝 У тебя нет целей для переключения")
	}

	// TODO: Реализовать inline кнопки для выбора цели
	return c.Send("🔄 Выбери цель для переключения:\n\n" + b.formatGoalsList(goals))
}

// handleText обрабатывает текстовые сообщения
func (b *Bot) handleText(c tele.Context) error {
	state := b.getOrCreateState(c.Sender().ID)
	text := c.Text()

	log.Printf("🔍 Текст пользователя: %s", text)
	log.Printf("🔍 Состояние пользователя: %s", state.State)

	switch state.State {
	case "waiting_goal_description":
		// Генерируем название цели через LLM
		title, err := b.llmClient.GenerateGoalTitle(text)
		if err != nil {
			return c.Send("❌ Ошибка при генерации названия цели")
		}

		// Создаем цель
		userID := strconv.FormatInt(c.Sender().ID, 10)
		goal := models.NewGoal(userID, title, text)

		if err := b.repo.CreateGoal(goal); err != nil {
			return c.Send("❌ Ошибка при создании цели")
		}

		// Устанавливаем как активную
		user, err := b.repo.GetUser(userID)
		if err != nil {
			return c.Send("❌ Ошибка при получении пользователя")
		}
		user.ActiveGoalID = goal.ID
		if err := b.repo.UpdateUser(user); err != nil {
			return c.Send("❌ Ошибка при обновлении пользователя")
		}
		log.Printf("🔍 Пользователь: %+v", user)
		// Сбрасываем состояние
		state.State = "idle"
		state.TempData = make(map[string]string)

		message := fmt.Sprintf("🎯 Цель создана!\n\n**Название:** %s\n**Описание:** %s\n\nИспользуй /next чтобы получить первый шаг", goal.Title, goal.Description)
		return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdown})

	case "rephrasing":
		userID := strconv.FormatInt(c.Sender().ID, 10)
		user, err := b.repo.GetUser(userID)
		if err != nil {
			return c.Send("❌ Ошибка при получении данных пользователя")
		}

		currentStep, err := b.repo.GetCurrentStep(user.ActiveGoalID)
		if err != nil {
			return c.Send("❌ Ошибка при получении текущего шага")
		}

		goal, err := b.repo.GetGoal(user.ActiveGoalID)
		if err != nil {
			return c.Send("❌ Ошибка при получении цели")
		}

		// Переформулируем шаг через LLM
		response, err := b.llmClient.RephraseStep(goal, currentStep, text)
		if err != nil {
			return c.Send("❌ Ошибка при переформулировке шага")
		}

		// Обновляем шаг
		currentStep.Text = response.Step
		currentStep.Rephrase(text)
		if err := b.repo.UpdateStep(currentStep); err != nil {
			return c.Send("❌ Ошибка при обновлении шага")
		}

		// Сбрасываем состояние
		state.State = "idle"
		state.TempData = make(map[string]string)

		message := fmt.Sprintf("🔄 Шаг переформулирован:\n\n%s", currentStep.Text)
		return c.Send(message)

	default:
		return c.Send("💡 Используй команды для работы с ботом. Напиши /help для справки")
	}
}

// getOrCreateState получает или создает состояние пользователя
func (b *Bot) getOrCreateState(userID int64) *UserState {
	if state, exists := b.states[userID]; exists {
		return state
	}

	state := &UserState{
		UserID:   userID,
		State:    "idle",
		TempData: make(map[string]string),
	}
	b.states[userID] = state
	return state
}

// formatGoalsList форматирует список целей для отображения
func (b *Bot) formatGoalsList(goals []*models.Goal) string {
	var result strings.Builder
	for i, goal := range goals {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, goal.Title))
	}
	return result.String()
}

// completeGoal завершает цель и сбрасывает активную цель пользователя
func (b *Bot) completeGoal(goal *models.Goal, user *models.User, completionReason string) error {
	// Отмечаем цель как завершенную
	goal.Status = "completed"
	now := time.Now()
	goal.CompletedAt = &now
	goal.UpdatedAt = now

	if err := b.repo.UpdateGoal(goal); err != nil {
		return fmt.Errorf("failed to update goal: %w", err)
	}

	// Сбрасываем активную цель пользователя
	user.ActiveGoalID = ""
	if err := b.repo.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// handleComplete обрабатывает команду /complete
func (b *Bot) handleComplete(c tele.Context) error {
	userID := strconv.FormatInt(c.Sender().ID, 10)

	user, err := b.repo.GetUser(userID)
	if err != nil {
		return c.Send("❌ Ошибка при получении данных пользователя")
	}

	if user.ActiveGoalID == "" {
		return c.Send("📝 У тебя нет активной цели")
	}

	goal, err := b.repo.GetGoal(user.ActiveGoalID)
	if err != nil {
		return c.Send("❌ Ошибка при получении цели")
	}

	// Отмечаем цель как завершенную
	goal.Status = "completed"
	now := time.Now()
	goal.CompletedAt = &now
	goal.UpdatedAt = now

	if err := b.repo.UpdateGoal(goal); err != nil {
		return c.Send("❌ Ошибка при обновлении цели")
	}

	// Сбрасываем активную цель пользователя
	user.ActiveGoalID = ""
	if err := b.repo.UpdateUser(user); err != nil {
		return c.Send("❌ Ошибка при обновлении пользователя")
	}

	message := fmt.Sprintf("🎉 **Поздравляю! Цель достигнута!**\n\n**%s**\n\nСоздай новую цель командой /newgoal",
		goal.Title)
	return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}
