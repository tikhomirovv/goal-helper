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
	State    string            // "idle", "waiting_goal_description", "rephrasing", "gathering_context"
	TempData map[string]string // Временные данные для создания цели
}

// NewBot создает нового бота
func NewBot(token string, repo repository.Repository, llmClient llm.Client) *Bot {
	// Настройки бота
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: BotPollerTimeout},
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
	b.bot.Handle(CmdStart, b.handleStart)
	b.bot.Handle(CmdHelp, b.handleHelp)
	b.bot.Handle(CmdGoals, b.handleGoals)
	b.bot.Handle(CmdNewGoal, b.handleNewGoal)
	b.bot.Handle(CmdStatus, b.handleStatus)
	b.bot.Handle(CmdStep, b.handleStep)
	b.bot.Handle(CmdDone, b.handleDone)
	b.bot.Handle(CmdNext, b.handleNext)
	b.bot.Handle(CmdRephrase, b.handleRephrase)
	b.bot.Handle(CmdSimpler, b.handleSimpler)
	b.bot.Handle(CmdSwitch, b.handleSwitch)
	b.bot.Handle(CmdComplete, b.handleComplete)
	b.bot.Handle(CmdContext, b.handleContext)

	// Обработчик кнопок
	b.bot.Handle(&tele.Btn{Text: BtnTextDone}, b.handleDone)
	b.bot.Handle(&tele.Btn{Text: BtnTextRephrase}, b.handleRephrase)
	b.bot.Handle(&tele.Btn{Text: BtnTextSimpler}, b.handleSimpler)
	b.bot.Handle(&tele.Btn{Text: BtnTextComplete}, b.handleComplete)

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
			return c.Send(MsgErrorCreateUser)
		}
	}

	// Создаем или обновляем состояние пользователя
	b.states[c.Sender().ID] = &UserState{
		UserID:   c.Sender().ID,
		State:    StateIdle,
		TempData: make(map[string]string),
	}

	// Приветственное сообщение
	message := fmt.Sprintf(MsgWelcomeTemplate, firstName)

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	btnGoals := menu.Text(BtnTextGoals)
	btnNewGoal := menu.Text(BtnTextNewGoal)

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
/simpler - Сделать текущий шаг проще (если он слишком сложный)
/complete - Завершить цель (если считаешь, что она достигнута)
/switch - Переключиться на другую цель
/context - Показать собранный контекст о тебе

**Как это работает:**
1. Создай цель командой /newgoal
2. Бот задаст несколько вопросов о твоем опыте и навыках
3. Получи первый шаг командой /next
4. Выполни шаг и отметь его командой /done
5. Получи следующий шаг командой /next
6. Повторяй, пока цель не будет достигнута

**Важно:** Каждый шаг должен быть максимально простым - от 5 минут до максимум 1 дня. Если шаг кажется слишком сложным, используй /simpler или /rephrase.

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
		return c.Send(MsgErrorSteps)
	}

	if len(goals) == 0 {
		return c.Send(MsgNoGoals)
	}

	var message strings.Builder
	message.WriteString("📋 **Твои цели:**\n\n")

	user, err := b.repo.GetUser(userID)
	if err != nil {
		return c.Send(MsgErrorUserData)
	}

	for i, goal := range goals {
		status := StatusIconInactive
		if goal.Status == GoalStatusCompleted {
			status = StatusIconCompleted
		} else if goal.ID == user.ActiveGoalID {
			status = StatusIconActive
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
	state.State = StateWaitingGoalDescription
	state.TempData = make(map[string]string)

	return c.Send(MsgNewGoalPrompt)
}

// handleStatus обрабатывает команду /status
func (b *Bot) handleStatus(c tele.Context) error {
	userID := strconv.FormatInt(c.Sender().ID, 10)

	user, err := b.repo.GetUser(userID)
	if err != nil {
		return c.Send(MsgErrorUserData)
	}

	if user.ActiveGoalID == "" {
		return c.Send(MsgNoActiveGoal)
	}

	goal, err := b.repo.GetGoal(user.ActiveGoalID)
	if err != nil {
		return c.Send(MsgErrorActiveGoal)
	}

	// Проверяем статус цели
	if goal.Status == GoalStatusCompleted {
		return c.Send(MsgGoalAlreadyCompleted)
	}

	steps, err := b.repo.GetGoalSteps(goal.ID)
	if err != nil {
		return c.Send(MsgErrorSteps)
	}

	completedCount := 0
	for _, step := range steps {
		if step.IsCompleted() {
			completedCount++
		}
	}

	message := fmt.Sprintf(MsgActiveGoalTemplate, goal.Title)
	if goal.Description != "" {
		message += fmt.Sprintf(MsgGoalDescriptionTemplate, goal.Description)
	}
	message += fmt.Sprintf(MsgProgressTemplate, completedCount, len(steps))
	message += MsgUseStepCommand

	return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

// handleStep обрабатывает команду /step
func (b *Bot) handleStep(c tele.Context) error {
	userID := strconv.FormatInt(c.Sender().ID, 10)

	user, err := b.repo.GetUser(userID)
	if err != nil {
		return c.Send(MsgErrorUserData)
	}

	if user.ActiveGoalID == "" {
		return c.Send(MsgNoActiveGoal)
	}

	goal, err := b.repo.GetGoal(user.ActiveGoalID)
	if err != nil {
		return c.Send(MsgErrorActiveGoal)
	}

	// Проверяем статус цели
	if goal.Status == GoalStatusCompleted {
		return c.Send(MsgGoalAlreadyCompleted)
	}

	currentStep, err := b.repo.GetCurrentStep(user.ActiveGoalID)
	if err != nil {
		return c.Send(MsgAllStepsCompleted)
	}

	message := fmt.Sprintf(MsgCurrentStepTemplate, currentStep.Text)

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	btnDone := menu.Text(BtnTextDone)
	btnRephrase := menu.Text(BtnTextRephrase)
	btnSimpler := menu.Text(BtnTextSimpler)

	menu.Reply(
		menu.Row(btnDone),
		menu.Row(btnRephrase, btnSimpler),
	)

	return c.Send(message, menu, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

// handleDone обрабатывает команду /done
func (b *Bot) handleDone(c tele.Context) error {
	userID := strconv.FormatInt(c.Sender().ID, 10)

	user, err := b.repo.GetUser(userID)
	if err != nil {
		return c.Send(MsgErrorUserData)
	}

	if user.ActiveGoalID == "" {
		return c.Send(MsgNoActiveGoal)
	}

	goal, err := b.repo.GetGoal(user.ActiveGoalID)
	if err != nil {
		return c.Send(MsgErrorGoal)
	}

	// Проверяем статус цели
	if goal.Status == GoalStatusCompleted {
		return c.Send(MsgGoalAlreadyCompleted)
	}

	currentStep, err := b.repo.GetCurrentStep(user.ActiveGoalID)
	if err != nil {
		return c.Send(MsgAllStepsCompleted)
	}

	// Отмечаем шаг как выполненный
	currentStep.Complete()
	if err := b.repo.UpdateStep(currentStep); err != nil {
		return c.Send(MsgErrorUpdateStep)
	}

	return c.Send(MsgStepCompleted)
}

// handleNext обрабатывает команду /next
func (b *Bot) handleNext(c tele.Context) error {
	userID := strconv.FormatInt(c.Sender().ID, 10)

	user, err := b.repo.GetUser(userID)
	if err != nil {
		return c.Send(MsgErrorUserData)
	}

	if user.ActiveGoalID == "" {
		return c.Send(MsgNoActiveGoal)
	}

	goal, err := b.repo.GetGoal(user.ActiveGoalID)
	if err != nil {
		return c.Send(MsgErrorGoal)
	}

	// Проверяем статус цели
	if goal.Status == GoalStatusCompleted {
		return c.Send(MsgGoalAlreadyCompleted)
	}

	// Получаем все шаги для цели
	allSteps, err := b.repo.GetGoalSteps(goal.ID)
	if err != nil {
		return c.Send(MsgErrorSteps)
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
		message := fmt.Sprintf(MsgUnfinishedStepTemplate, currentStep.Text)
		return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	// Все шаги выполнены, генерируем следующий
	log.Printf("🔍 Генерируем следующий шаг для цели: %s", goal.Title)
	log.Printf("🔍 Количество выполненных шагов: %d", len(completedSteps))

	// Сначала проверяем, нужен ли сбор контекста
	if len(completedSteps) == 0 && len(goal.Context.Clarifications) == 0 {
		// Это первый шаг и контекст не собран - собираем контекст
		log.Printf("🔍 Собираем контекст для новой цели")
		contextResponse, err := b.llmClient.GatherContext(goal)
		if err != nil {
			log.Printf("❌ Ошибка при сборе контекста: %v", err)
			return c.Send(MsgErrorGatherContext)
		}

		if contextResponse.Status == LLMStatusNeedContext {
			// Нужен дополнительный контекст
			state := b.getOrCreateState(c.Sender().ID)
			state.State = StateGatheringContext
			state.TempData["goal_id"] = goal.ID
			state.TempData["context_question"] = contextResponse.Question

			message := fmt.Sprintf(MsgContextQuestionTemplate, contextResponse.Question)
			return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
		}
	}

	response, err := b.llmClient.GenerateStep(goal, completedSteps)
	if err != nil {
		log.Printf("❌ Ошибка при генерации шага: %v", err)
		return c.Send(fmt.Sprintf("❌ Ошибка при генерации шага: %v", err))
	}

	log.Printf("🔍 Получен ответ от LLM: статус=%s, шаг=%s", response.Status, response.Step)

	if response.Status == LLMStatusNeedClarification {
		return c.Send(fmt.Sprintf(MsgClarificationTemplate, response.Question))
	}

	// Обрабатываем завершение цели
	if response.Status == LLMStatusGoalCompleted {
		// Получаем пользователя для завершения цели
		userID := strconv.FormatInt(c.Sender().ID, 10)
		user, err := b.repo.GetUser(userID)
		if err != nil {
			return c.Send(MsgErrorUserData)
		}

		if err := b.completeGoal(goal, user, response.CompletionReason); err != nil {
			return c.Send(MsgErrorUpdateGoal)
		}
		message := fmt.Sprintf(MsgGoalCompletedTemplate, goal.Title, response.CompletionReason)
		return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	// Обрабатываем близость к завершению
	if response.Status == LLMStatusNearCompletion {
		// Создаем новый шаг
		newStep := models.NewStep(goal.ID, response.Step)
		if err := b.repo.CreateStep(newStep); err != nil {
			return c.Send(MsgErrorCreateStep)
		}

		message := fmt.Sprintf(MsgNearCompletionTemplate, newStep.Text)

		menu := &tele.ReplyMarkup{ResizeKeyboard: true}
		btnDone := menu.Text(BtnTextDone)
		btnRephrase := menu.Text(BtnTextRephrase)
		btnSimpler := menu.Text(BtnTextSimpler)
		btnComplete := menu.Text(BtnTextComplete)

		menu.Reply(
			menu.Row(btnDone),
			menu.Row(btnRephrase, btnSimpler),
			menu.Row(btnComplete),
		)

		return c.Send(message, menu, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	// Обычный шаг
	if response.Status == LLMStatusOK {
		// Создаем новый шаг
		newStep := models.NewStep(goal.ID, response.Step)
		if err := b.repo.CreateStep(newStep); err != nil {
			return c.Send(MsgErrorCreateStep)
		}

		message := fmt.Sprintf(MsgNewStepTemplate, newStep.Text)

		menu := &tele.ReplyMarkup{ResizeKeyboard: true}
		btnDone := menu.Text(BtnTextDone)
		btnRephrase := menu.Text(BtnTextRephrase)
		btnSimpler := menu.Text(BtnTextSimpler)

		menu.Reply(
			menu.Row(btnDone),
			menu.Row(btnRephrase, btnSimpler),
		)

		return c.Send(message, menu, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
	}

	// Неизвестный статус
	return c.Send(MsgErrorUnexpectedResponse)
}

// handleRephrase обрабатывает команду /rephrase
func (b *Bot) handleRephrase(c tele.Context) error {
	// Устанавливаем состояние "переформулировка"
	state := b.getOrCreateState(c.Sender().ID)
	state.State = StateRephrasing
	state.TempData = make(map[string]string)

	return c.Send(MsgRephrasePrompt)
}

// handleSimpler обрабатывает команду /simpler
func (b *Bot) handleSimpler(c tele.Context) error {
	userID := strconv.FormatInt(c.Sender().ID, 10)

	user, err := b.repo.GetUser(userID)
	if err != nil {
		return c.Send(MsgErrorUserData)
	}

	if user.ActiveGoalID == "" {
		return c.Send(MsgNoActiveGoal)
	}

	goal, err := b.repo.GetGoal(user.ActiveGoalID)
	if err != nil {
		return c.Send(MsgErrorGoal)
	}

	// Проверяем статус цели
	if goal.Status == GoalStatusCompleted {
		return c.Send(MsgGoalAlreadyCompleted)
	}

	currentStep, err := b.repo.GetCurrentStep(user.ActiveGoalID)
	if err != nil {
		return c.Send(MsgAllStepsCompleted)
	}

	// Переформулируем шаг с просьбой сделать его проще
	response, err := b.llmClient.RephraseStep(goal, currentStep, MsgSimplifyPrompt)
	if err != nil {
		return c.Send(MsgErrorSimplifyStep)
	}

	// Обновляем шаг
	currentStep.Text = response.Step
	currentStep.Rephrase(MsgUserRequestedSimplification)
	if err := b.repo.UpdateStep(currentStep); err != nil {
		return c.Send(MsgErrorUpdateStep)
	}

	message := fmt.Sprintf(MsgStepSimplifiedTemplate, currentStep.Text)

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	btnDone := menu.Text(BtnTextDone)
	btnRephrase := menu.Text(BtnTextRephrase)

	menu.Reply(
		menu.Row(btnDone),
		menu.Row(btnRephrase),
	)

	return c.Send(message, menu, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

// handleSwitch обрабатывает команду /switch
func (b *Bot) handleSwitch(c tele.Context) error {
	userID := strconv.FormatInt(c.Sender().ID, 10)

	goals, err := b.repo.GetUserGoals(userID)
	if err != nil {
		return c.Send(MsgErrorSteps)
	}

	if len(goals) == 0 {
		return c.Send(MsgNoGoalsForSwitch)
	}

	// TODO: Реализовать inline кнопки для выбора цели
	return c.Send(MsgSwitchGoalsPrompt + b.formatGoalsList(goals))
}

// handleText обрабатывает текстовые сообщения
func (b *Bot) handleText(c tele.Context) error {
	state := b.getOrCreateState(c.Sender().ID)
	text := c.Text()

	log.Printf("🔍 Текст пользователя: %s", text)
	log.Printf("🔍 Состояние пользователя: %s", state.State)

	switch state.State {
	case StateWaitingGoalDescription:
		// Генерируем название цели через LLM
		title, err := b.llmClient.GenerateGoalTitle(text)
		if err != nil {
			return c.Send(MsgErrorGenerateStep)
		}

		// Создаем цель
		userID := strconv.FormatInt(c.Sender().ID, 10)
		goal := models.NewGoal(userID, title, text)

		if err := b.repo.CreateGoal(goal); err != nil {
			return c.Send(MsgErrorCreateGoal)
		}

		// Устанавливаем как активную
		user, err := b.repo.GetUser(userID)
		if err != nil {
			return c.Send(MsgErrorUserData)
		}
		user.ActiveGoalID = goal.ID
		if err := b.repo.UpdateUser(user); err != nil {
			return c.Send(MsgErrorUpdateUser)
		}
		log.Printf("🔍 Пользователь: %+v", user)
		// Сбрасываем состояние
		state.State = StateIdle
		state.TempData = make(map[string]string)

		message := fmt.Sprintf(MsgGoalCreatedTemplate, goal.Title, goal.Description)
		return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdown})

	case StateGatheringContext:
		// Обрабатываем ответ на вопрос о контексте
		goalID := state.TempData["goal_id"]
		question := state.TempData["context_question"]

		if goalID == "" {
			return c.Send(MsgGoalNotFoundError)
		}

		// Получаем цель
		goal, err := b.repo.GetGoal(goalID)
		if err != nil {
			return c.Send(MsgErrorGoal)
		}

		// Добавляем уточнение в контекст
		goal.AddClarification(question, text)
		if err := b.repo.UpdateGoal(goal); err != nil {
			return c.Send(MsgErrorUpdateGoal)
		}

		// Проверяем, нужен ли еще контекст
		contextResponse, err := b.llmClient.GatherContext(goal)
		if err != nil {
			return c.Send(MsgErrorGatherContext)
		}

		if contextResponse.Status == LLMStatusNeedContext {
			// Нужен еще контекст
			state.TempData["context_question"] = contextResponse.Question
			message := fmt.Sprintf(MsgContextThanksTemplate, contextResponse.Question)
			return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
		}

		// Контекст собран, генерируем первый шаг
		completedSteps := []*models.Step{} // Пустой массив для первого шага
		response, err := b.llmClient.GenerateStep(goal, completedSteps)
		if err != nil {
			return c.Send(MsgErrorGenerateStep)
		}

		// Сбрасываем состояние
		state.State = StateIdle
		state.TempData = make(map[string]string)

		// Обрабатываем ответ LLM
		if response.Status == LLMStatusNeedClarification {
			return c.Send(fmt.Sprintf(MsgClarificationTemplate, response.Question))
		}

		if response.Status == LLMStatusGoalCompleted {
			// Получаем пользователя для завершения цели
			userID := strconv.FormatInt(c.Sender().ID, 10)
			user, err := b.repo.GetUser(userID)
			if err != nil {
				return c.Send(MsgErrorUserData)
			}

			if err := b.completeGoal(goal, user, response.CompletionReason); err != nil {
				return c.Send(MsgErrorUpdateGoal)
			}
			message := fmt.Sprintf(MsgGoalCompletedTemplate, goal.Title, response.CompletionReason)
			return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
		}

		if response.Status == LLMStatusNearCompletion {
			// Создаем новый шаг
			newStep := models.NewStep(goal.ID, response.Step)
			if err := b.repo.CreateStep(newStep); err != nil {
				return c.Send(MsgErrorCreateStep)
			}

			message := fmt.Sprintf(MsgNearCompletionTemplate, newStep.Text)

			menu := &tele.ReplyMarkup{ResizeKeyboard: true}
			btnDone := menu.Text(BtnTextDone)
			btnRephrase := menu.Text(BtnTextRephrase)
			btnSimpler := menu.Text(BtnTextSimpler)
			btnComplete := menu.Text(BtnTextComplete)

			menu.Reply(
				menu.Row(btnDone),
				menu.Row(btnRephrase, btnSimpler),
				menu.Row(btnComplete),
			)

			return c.Send(message, menu, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
		}

		if response.Status == LLMStatusOK {
			// Создаем новый шаг
			newStep := models.NewStep(goal.ID, response.Step)
			if err := b.repo.CreateStep(newStep); err != nil {
				return c.Send(MsgErrorCreateStep)
			}

			message := fmt.Sprintf(MsgFirstStepTemplate, newStep.Text)

			menu := &tele.ReplyMarkup{ResizeKeyboard: true}
			btnDone := menu.Text(BtnTextDone)
			btnRephrase := menu.Text(BtnTextRephrase)
			btnSimpler := menu.Text(BtnTextSimpler)

			menu.Reply(
				menu.Row(btnDone),
				menu.Row(btnRephrase, btnSimpler),
			)

			return c.Send(message, menu, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
		}

		return c.Send(MsgErrorUnexpectedResponse)

	case StateRephrasing:
		userID := strconv.FormatInt(c.Sender().ID, 10)
		user, err := b.repo.GetUser(userID)
		if err != nil {
			return c.Send(MsgErrorUserData)
		}

		currentStep, err := b.repo.GetCurrentStep(user.ActiveGoalID)
		if err != nil {
			return c.Send(MsgCurrentStepError)
		}

		goal, err := b.repo.GetGoal(user.ActiveGoalID)
		if err != nil {
			return c.Send(MsgErrorGoal)
		}

		// Переформулируем шаг через LLM
		response, err := b.llmClient.RephraseStep(goal, currentStep, text)
		if err != nil {
			return c.Send(MsgErrorRephraseStep)
		}

		// Обновляем шаг
		currentStep.Text = response.Step
		currentStep.Rephrase(text)
		if err := b.repo.UpdateStep(currentStep); err != nil {
			return c.Send(MsgErrorUpdateStep)
		}

		// Сбрасываем состояние
		state.State = StateIdle
		state.TempData = make(map[string]string)

		message := fmt.Sprintf(MsgStepRephrasedTemplate, currentStep.Text)
		return c.Send(message)

	default:
		return c.Send(MsgHelpDefault)
	}
}

// getOrCreateState получает или создает состояние пользователя
func (b *Bot) getOrCreateState(userID int64) *UserState {
	if state, exists := b.states[userID]; exists {
		return state
	}

	state := &UserState{
		UserID:   userID,
		State:    StateIdle,
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
	goal.Status = GoalStatusCompleted
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
		return c.Send(MsgErrorUserData)
	}

	if user.ActiveGoalID == "" {
		return c.Send(MsgNoActiveGoal)
	}

	goal, err := b.repo.GetGoal(user.ActiveGoalID)
	if err != nil {
		return c.Send(MsgErrorGoal)
	}

	// Отмечаем цель как завершенную
	goal.Status = GoalStatusCompleted
	now := time.Now()
	goal.CompletedAt = &now
	goal.UpdatedAt = now

	if err := b.repo.UpdateGoal(goal); err != nil {
		return c.Send(MsgErrorUpdateGoal)
	}

	// Сбрасываем активную цель пользователя
	user.ActiveGoalID = ""
	if err := b.repo.UpdateUser(user); err != nil {
		return c.Send(MsgErrorUpdateUser)
	}

	message := fmt.Sprintf(MsgGoalCompletedManualTemplate, goal.Title)
	return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}

// handleContext обрабатывает команду /context
func (b *Bot) handleContext(c tele.Context) error {
	userID := strconv.FormatInt(c.Sender().ID, 10)

	user, err := b.repo.GetUser(userID)
	if err != nil {
		return c.Send(MsgErrorUserData)
	}

	if user.ActiveGoalID == "" {
		return c.Send(MsgNoActiveGoal)
	}

	goal, err := b.repo.GetGoal(user.ActiveGoalID)
	if err != nil {
		return c.Send(MsgErrorGoal)
	}

	message := fmt.Sprintf(MsgContextSummaryTemplate, goal.Title)
	message += goal.GetContextSummary()

	return c.Send(message, &tele.SendOptions{ParseMode: tele.ModeMarkdown})
}
