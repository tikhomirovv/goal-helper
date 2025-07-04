# 📌 Идея проекта: Помощник в достижении целей через микро-задачи

## 🔹 Суть
Проект — это не планировщик в классическом смысле, а **интерактивный помощник по достижению целей**, который помогает пользователю пройти путь к результату **через минимальные, неотталкивающие шаги**.

Идея вдохновлена логикой видеоигр и нарративных систем:
- Пользователь **не видит сразу весь план**,
- Получает **одну задачу за раз**,
- Не знает, **сколько шагов ещё впереди**,
- Каждое задание — простое, достижимое, не вызывает сопротивления.

## 🔹 Основные принципы

1. **Одна цель = путь, разбитый на микро-шаги**
2. **Шаги минимальны** (время выполнения может быть от 5 минут до пары часов)
3. **Один шаг — один логический элемент действия**
4. **Следующий шаг выдаётся только после выполнения текущего**
5. **Пользователь может переформулировать шаг, если он не подходит**
6. **Нет привязки ко времени (не "на день"), а к смысловой завершённости**

## 🔹 Работа LLM в системе

### ✅ Генерация шагов
LLM получает:
- Цель пользователя (и, возможно, её описание)
- Выполненные ранее шаги (история)
- Контекст (уточняющие ответы пользователя, если были)

И возвращает:
- `step`: следующий микро-шаг

### ✅ Если LLM не хватает данных
LLM должна вернуть:
```json
{
  "status": "need_clarification",
  "question": "У тебя уже есть идея для темы песни?"
}
```

### ✅ При переформулировке шага
Пользователь может отклонить текущий шаг и оставить комментарий, например:
> "Этот шаг слишком общий, не понимаю, что делать"

LLM получает:
- цель,
- историю,
- текущий шаг,
- комментарий,
и возвращает альтернативную формулировку.

## 🔹 Типовая структура целей и шагов в БД

```json
Goal {
  id: uuid,
  user_id: uuid,
  title: string,
  description: string,
  current_step: Step,
  completed_steps: Step[],
  context: {
    "clarifications": [ ... ]
  }
}

Step {
  id: uuid,
  goal_id: uuid,
  text: string,
  created_at: timestamp,
  completed_at: timestamp | null,
  rephrased: boolean,
  user_comment: string | null
}
```

## 🔹 Telegram-бот как интерфейс MVP

### ✅ Почему Telegram:
- Не требует фронтенда
- Поддерживает состояние через FSM
- Идеально подходит для пошагового взаимодействия
- Интуитивен для пользователя

## 🔹 FSM (Finite State Machine) — описание

FSM = конечный автомат. У бота есть состояния (например, «нет цели», «работа над целью»), между которыми он переходит в зависимости от ввода пользователя.

## 🔹 Основной пользовательский flow с активной целью

### 1. `/start` — Приветствие
- Кнопки: 📋 Список целей, ➕ Новая цель

### 2. `/goals` — Список целей
- Показываются все цели пользователя:
  ```
  1. 🎯 Выучить японский
  2. 🎼 Написать альбом
  ```
- Выбор цели делает её активной

### 3. `/newgoal` — Создание цели
- Бот запрашивает название и описание
- Устанавливает её как активную

### 4. Работа с активной целью
- `/step` — текущий шаг
- `/done` — отметить шаг выполненным
- `/next` — получить следующий шаг
- `/rephrase` — переформулировать шаг

### 5. `/switch` — Смена активной цели
- Показывает список целей

### 6. `/status` — Показать активную цель и прогресс
### 7. `/help` — Справка по командам

## 🔹 Промпты для LLM

### ✅ 1. Генерация первого шага
```plaintext
Ты коуч, помогаешь пользователю достичь цели, разбивая её на минимальные, простые задачи.

Если тебе не хватает данных — задай уточняющий вопрос.

Формат:
{
  "status": "ok" | "need_clarification",
  "step": "...",
  "question": "..."
}
```

### ✅ 2. Генерация следующего шага (после выполнения предыдущего)
```plaintext
Цель: <<>>
История шагов:
- ...
- ...
Сгенерируй следующий логичный шаг. Если не уверен — задай вопрос.
```

### ✅ 3. Переформулировка шага
```plaintext
Цель: <<>>
История: ...
Текущий шаг: <<...>>
Комментарий пользователя: <<...>>

Сформулируй альтернативный шаг на том же уровне сложности.
```

### ✅ 4. Запрос уточнения при вводе цели
```plaintext
Цель: <<>>

Если цель недостаточно понятна для генерации шага — верни статус и вопрос:
{
  "status": "need_clarification",
  "question": "..."
}
```

## 🧠 Дополнительные возможности в будущем:

- Классификация целей по типам (творческие, образовательные, бытовые и т.д.)
- Характер шагов и глубина — настраиваются по типу
- Поддержка нескольких активных целей через кнопки
- Возможность видеть прогресс (например, "5 шагов из 20")
- Интеграция локальной LLM через абстракцию
- Настройка «тона» помощника (серьёзный, дружелюбный, мотивационный)
