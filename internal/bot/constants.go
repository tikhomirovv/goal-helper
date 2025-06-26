package bot

// Константы для состояний пользователя
const (
	StateIdle                   = "idle"
	StateWaitingGoalDescription = "waiting_goal_description"
	StateRephrasing             = "rephrasing"
	StateGatheringContext       = "gathering_context"
)

// Константы для статусов целей
const (
	GoalStatusActive    = "active"
	GoalStatusCompleted = "completed"
	GoalStatusInactive  = "inactive"
)

// Константы для статусов ответов LLM
const (
	LLMStatusOK                = "ok"
	LLMStatusNeedClarification = "need_clarification"
	LLMStatusGoalCompleted     = "goal_completed"
	LLMStatusNearCompletion    = "near_completion"
	LLMStatusNeedContext       = "need_context"
)

// Константы для сообщений
const (
	MsgErrorUserData           = "❌ Ошибка при получении данных пользователя"
	MsgErrorGoal               = "❌ Ошибка при получении цели"
	MsgErrorActiveGoal         = "❌ Ошибка при получении активной цели"
	MsgErrorSteps              = "❌ Ошибка при получении шагов"
	MsgErrorCreateGoal         = "❌ Ошибка при создании цели"
	MsgErrorCreateUser         = "❌ Ошибка при создании пользователя"
	MsgErrorUpdateGoal         = "❌ Ошибка при обновлении цели"
	MsgErrorUpdateUser         = "❌ Ошибка при обновлении пользователя"
	MsgErrorUpdateStep         = "❌ Ошибка при обновлении шага"
	MsgErrorCreateStep         = "❌ Ошибка при создании шага"
	MsgErrorGenerateStep       = "❌ Ошибка при генерации шага"
	MsgErrorGatherContext      = "❌ Ошибка при сборе контекста"
	MsgErrorRephraseStep       = "❌ Ошибка при переформулировке шага"
	MsgErrorSimplifyStep       = "❌ Ошибка при упрощении шага"
	MsgErrorUnexpectedResponse = "❌ Неожиданный ответ от системы"
)

// Константы для статусов целей в UI
const (
	StatusIconActive    = "🎯"
	StatusIconCompleted = "✅"
	StatusIconInactive  = "⏳"
)

// Константы для кнопок
const (
	BtnTextDone     = "✅ Выполнил"
	BtnTextRephrase = "🔄 Переформулировать"
	BtnTextSimpler  = "🔽 Упростить"
	BtnTextComplete = "🎉 Завершить цель"
	BtnTextGoals    = "📋 Мои цели"
	BtnTextNewGoal  = "➕ Новая цель"
)

// Константы для команд
const (
	CmdStart    = "/start"
	CmdHelp     = "/help"
	CmdGoals    = "/goals"
	CmdNewGoal  = "/newgoal"
	CmdStatus   = "/status"
	CmdStep     = "/step"
	CmdDone     = "/done"
	CmdNext     = "/next"
	CmdRephrase = "/rephrase"
	CmdSimpler  = "/simpler"
	CmdSwitch   = "/switch"
	CmdComplete = "/complete"
	CmdContext  = "/context"
)

// Константы для сообщений пользователю
const (
	MsgWelcomeTemplate             = "🎯 Привет, %s!\n\nЯ помогу тебе достичь целей через простые шаги.\n\nЧто хочешь сделать?"
	MsgNoGoals                     = "📝 У тебя пока нет целей.\n\nСоздай первую цель командой /newgoal"
	MsgNoActiveGoal                = "📝 У тебя нет активной цели.\n\nВыбери цель из списка командой /goals или создай новую командой /newgoal"
	MsgGoalAlreadyCompleted        = "✅ Эта цель уже завершена!\n\nСоздай новую цель командой /newgoal или выбери другую из списка /goals"
	MsgAllStepsCompleted           = "✅ Поздравляю! Ты выполнил все шаги для этой цели.\n\nИспользуй /next чтобы получить следующий шаг"
	MsgStepCompleted               = "✅ Отлично! Шаг выполнен.\n\nИспользуй /next чтобы получить следующий шаг"
	MsgGoalCreatedTemplate         = "🎯 Цель создана!\n\n**Название:** %s\n**Описание:** %s\n\nИспользуй /next чтобы получить первый шаг"
	MsgGoalCompletedTemplate       = "🎉 **Поздравляю! Цель достигнута!**\n\n**%s**\n\n%s\n\nСоздай новую цель командой /newgoal"
	MsgNearCompletionTemplate      = "🎯 **Почти готово! Осталось совсем немного:**\n\n%s\n\n💡 После этого шага цель может быть достигнута!"
	MsgStepSimplifiedTemplate      = "🔄 Шаг упрощен:\n\n**%s**\n\n💡 Теперь этот шаг должен быть намного проще!"
	MsgStepRephrasedTemplate       = "🔄 Шаг переформулирован:\n\n%s"
	MsgContextQuestionTemplate     = "🔍 Для более точной помощи мне нужно узнать немного больше о тебе:\n\n**%s**\n\nОтветь на этот вопрос, и я смогу предложить подходящие шаги."
	MsgContextThanksTemplate       = "🔍 Спасибо! Теперь еще один вопрос:\n\n**%s**"
	MsgRephrasePrompt              = "🔄 Опиши, что именно не подходит в текущем шаге?\n\nНапример: \"Слишком сложно\", \"Непонятно что делать\", \"Нужно что-то проще\""
	MsgHelpDefault                 = "💡 Используй команды для работы с ботом. Напиши /help для справки"
	MsgNoGoalsForSwitch            = "📝 У тебя нет целей для переключения"
	MsgSwitchGoalsPrompt           = "🔄 Выбери цель для переключения:\n\n"
	MsgGoalNotFoundError           = "❌ Ошибка: не найден ID цели"
	MsgCurrentStepError            = "❌ Ошибка при получении текущего шага"
	MsgGoalCompletedManualTemplate = "🎉 **Поздравляю! Цель достигнута!**\n\n**%s**\n\nСоздай новую цель командой /newgoal"
	MsgNewGoalPrompt               = "🎯 Отлично! Давай создадим новую цель.\n\nОпиши свою цель подробно - что именно ты хочешь достичь? Я сам придумаю подходящее название."
	MsgActiveGoalTemplate          = "🎯 **Активная цель:** %s\n\n"
	MsgGoalDescriptionTemplate     = "📝 %s\n\n"
	MsgProgressTemplate            = "📊 **Прогресс:** %d/%d шагов выполнено\n\n"
	MsgUseStepCommand              = "Используй /step чтобы увидеть текущий шаг"
	MsgCurrentStepTemplate         = "📝 **Текущий шаг:**\n\n%s"
	MsgUnfinishedStepTemplate      = "⏳ У тебя есть невыполненный шаг:\n\n**%s**\n\nСначала выполни этот шаг командой /done, а потом получи следующий."
	MsgClarificationTemplate       = "❓ %s"
	MsgNewStepTemplate             = "📝 **Новый шаг:**\n\n%s"
	MsgFirstStepTemplate           = "📝 **Первый шаг:**\n\n%s"
	MsgContextSummaryTemplate      = "📋 **Контекст для цели:** %s\n\n"
	MsgSimplifyPrompt              = "Сделай этот шаг максимально простым - от 5 минут до максимум 1 дня. Разбей на самую простую возможную задачу."
	MsgUserRequestedSimplification = "Пользователь запросил упрощение шага"
)

// Константы для настройки бота
const (
	BotPollerTimeout = 10
)
