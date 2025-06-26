package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"goal-helper/internal/models"
)

// FileRepository реализует Repository интерфейс через JSON файлы
type FileRepository struct {
	dataDir string
	mutex   sync.RWMutex

	// Кэш данных в памяти для быстрого доступа
	users map[string]*models.User
	goals map[string]*models.Goal
	steps map[string]*models.Step
}

// NewFileRepository создает новый файловый репозиторий
func NewFileRepository(dataDir string) (*FileRepository, error) {
	repo := &FileRepository{
		dataDir: dataDir,
		users:   make(map[string]*models.User),
		goals:   make(map[string]*models.Goal),
		steps:   make(map[string]*models.Step),
	}

	// Создаем директорию, если её нет
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Загружаем существующие данные
	if err := repo.loadData(); err != nil {
		return nil, fmt.Errorf("failed to load data: %w", err)
	}

	return repo, nil
}

// loadData загружает все данные из файлов в память
func (r *FileRepository) loadData() error {
	// Загружаем пользователей
	if err := r.loadUsers(); err != nil {
		return err
	}

	// Загружаем цели
	if err := r.loadGoals(); err != nil {
		return err
	}

	// Загружаем шаги
	if err := r.loadSteps(); err != nil {
		return err
	}

	return nil
}

// loadUsers загружает пользователей из файла
func (r *FileRepository) loadUsers() error {
	filename := filepath.Join(r.dataDir, "users.json")

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Файл не существует, создаем пустой
			return r.saveUsers()
		}
		return err
	}

	var users []*models.User
	if err := json.Unmarshal(data, &users); err != nil {
		return err
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, user := range users {
		r.users[user.ID] = user
	}

	return nil
}

// saveUsers сохраняет пользователей в файл
func (r *FileRepository) saveUsers() error {

	r.mutex.RLock()
	users := make([]*models.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	r.mutex.RUnlock()

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(r.dataDir, "users.json")

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// loadGoals загружает цели из файла
func (r *FileRepository) loadGoals() error {
	filename := filepath.Join(r.dataDir, "goals.json")

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return r.saveGoals()
		}
		return err
	}

	var goals []*models.Goal
	if err := json.Unmarshal(data, &goals); err != nil {
		return err
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, goal := range goals {
		r.goals[goal.ID] = goal
	}

	return nil
}

// saveGoals сохраняет цели в файл
func (r *FileRepository) saveGoals() error {
	r.mutex.RLock()
	goals := make([]*models.Goal, 0, len(r.goals))
	for _, goal := range r.goals {
		goals = append(goals, goal)
	}
	r.mutex.RUnlock()

	data, err := json.MarshalIndent(goals, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(r.dataDir, "goals.json")
	return os.WriteFile(filename, data, 0644)
}

// loadSteps загружает шаги из файла
func (r *FileRepository) loadSteps() error {
	filename := filepath.Join(r.dataDir, "steps.json")

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return r.saveSteps()
		}
		return err
	}

	var steps []*models.Step
	if err := json.Unmarshal(data, &steps); err != nil {
		return err
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, step := range steps {
		r.steps[step.ID] = step
	}

	return nil
}

// saveSteps сохраняет шаги в файл
func (r *FileRepository) saveSteps() error {
	r.mutex.RLock()
	steps := make([]*models.Step, 0, len(r.steps))
	for _, step := range r.steps {
		steps = append(steps, step)
	}
	r.mutex.RUnlock()

	data, err := json.MarshalIndent(steps, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(r.dataDir, "steps.json")
	return os.WriteFile(filename, data, 0644)
}

// Реализация методов интерфейса Repository
func (r *FileRepository) GetUser(userID string) (*models.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	return user, nil
}

func (r *FileRepository) CreateUser(user *models.User) error {
	r.mutex.Lock()

	if _, exists := r.users[user.ID]; exists {
		r.mutex.Unlock()
		return fmt.Errorf("user already exists: %s", user.ID)
	}

	r.users[user.ID] = user
	r.mutex.Unlock()

	return r.saveUsers()
}

func (r *FileRepository) UpdateUser(user *models.User) error {
	r.mutex.Lock()

	if _, exists := r.users[user.ID]; !exists {
		r.mutex.Unlock()
		return fmt.Errorf("user not found: %s", user.ID)
	}

	r.users[user.ID] = user
	r.mutex.Unlock()

	return r.saveUsers()
}

func (r *FileRepository) GetGoal(goalID string) (*models.Goal, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	goal, exists := r.goals[goalID]
	if !exists {
		return nil, fmt.Errorf("goal not found: %s", goalID)
	}

	return goal, nil
}

func (r *FileRepository) GetUserGoals(userID string) ([]*models.Goal, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var userGoals []*models.Goal
	for _, goal := range r.goals {
		if goal.UserID == userID {
			userGoals = append(userGoals, goal)
		}
	}

	return userGoals, nil
}

func (r *FileRepository) CreateGoal(goal *models.Goal) error {
	r.mutex.Lock()

	if _, exists := r.goals[goal.ID]; exists {
		r.mutex.Unlock()
		return fmt.Errorf("goal already exists: %s", goal.ID)
	}

	r.goals[goal.ID] = goal
	r.mutex.Unlock()

	return r.saveGoals()
}

func (r *FileRepository) UpdateGoal(goal *models.Goal) error {
	r.mutex.Lock()

	if _, exists := r.goals[goal.ID]; !exists {
		r.mutex.Unlock()
		return fmt.Errorf("goal not found: %s", goal.ID)
	}

	goal.UpdatedAt = time.Now()
	r.goals[goal.ID] = goal
	r.mutex.Unlock()

	return r.saveGoals()
}

func (r *FileRepository) DeleteGoal(goalID string) error {
	r.mutex.Lock()

	if _, exists := r.goals[goalID]; !exists {
		r.mutex.Unlock()
		return fmt.Errorf("goal not found: %s", goalID)
	}

	delete(r.goals, goalID)

	// Удаляем все шаги этой цели
	for stepID, step := range r.steps {
		if step.GoalID == goalID {
			delete(r.steps, stepID)
		}
	}

	r.mutex.Unlock()

	if err := r.saveGoals(); err != nil {
		return err
	}

	return r.saveSteps()
}

func (r *FileRepository) GetStep(stepID string) (*models.Step, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	step, exists := r.steps[stepID]
	if !exists {
		return nil, fmt.Errorf("step not found: %s", stepID)
	}

	return step, nil
}

func (r *FileRepository) GetGoalSteps(goalID string) ([]*models.Step, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var goalSteps []*models.Step
	for _, step := range r.steps {
		if step.GoalID == goalID {
			goalSteps = append(goalSteps, step)
		}
	}

	// Сортируем шаги по дате создания (от старых к новым)
	sort.Slice(goalSteps, func(i, j int) bool {
		return goalSteps[i].CreatedAt.Before(goalSteps[j].CreatedAt)
	})

	return goalSteps, nil
}

func (r *FileRepository) GetCurrentStep(goalID string) (*models.Step, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Ищем последний невыполненный шаг для цели
	var currentStep *models.Step
	for _, step := range r.steps {
		if step.GoalID == goalID && !step.IsCompleted() {
			if currentStep == nil || step.CreatedAt.After(currentStep.CreatedAt) {
				currentStep = step
			}
		}
	}

	if currentStep == nil {
		return nil, fmt.Errorf("no current step found for goal: %s", goalID)
	}

	return currentStep, nil
}

func (r *FileRepository) CreateStep(step *models.Step) error {
	r.mutex.Lock()

	if _, exists := r.steps[step.ID]; exists {
		r.mutex.Unlock()
		return fmt.Errorf("step already exists: %s", step.ID)
	}

	r.steps[step.ID] = step
	r.mutex.Unlock()

	return r.saveSteps()
}

func (r *FileRepository) UpdateStep(step *models.Step) error {
	r.mutex.Lock()

	if _, exists := r.steps[step.ID]; !exists {
		r.mutex.Unlock()
		return fmt.Errorf("step not found: %s", step.ID)
	}

	r.steps[step.ID] = step
	r.mutex.Unlock()

	return r.saveSteps()
}

func (r *FileRepository) DeleteStep(stepID string) error {
	r.mutex.Lock()

	if _, exists := r.steps[stepID]; !exists {
		r.mutex.Unlock()
		return fmt.Errorf("step not found: %s", stepID)
	}

	delete(r.steps, stepID)
	r.mutex.Unlock()

	return r.saveSteps()
}

func (r *FileRepository) Close() error {
	// Сохраняем все данные перед закрытием
	if err := r.saveUsers(); err != nil {
		return err
	}
	if err := r.saveGoals(); err != nil {
		return err
	}
	if err := r.saveSteps(); err != nil {
		return err
	}

	return nil
}
