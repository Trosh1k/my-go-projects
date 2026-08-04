package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

var db *sql.DB

// getEnv возвращает значение переменной окружения или значение по умолчанию
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Получить все задачи
func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, description, done FROM tasks")
	if err != nil {
		log.WithError(err).Error("Ошибка БД")
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Done)
		if err != nil {
			log.WithError(err).Error("Ошибка чтения данных")
			http.Error(w, "Ошибка чтения данных", http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, t)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// Создать задачу
func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	var newTask Task
	if err := json.NewDecoder(r.Body).Decode(&newTask); err != nil {
		log.WithError(err).Error("Invalid JSON")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := db.Exec(
		"INSERT INTO tasks (title, description, done) VALUES (?, ?, ?)",
		newTask.Title, newTask.Description, false,
	)
	if err != nil {
		log.WithError(err).Error("Ошибка сохранения")
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}
	id, _ := result.LastInsertId()
	newTask.ID = int(id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTask)
}

// Получить задачу по ID
func getTaskByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/tasks/")
	var t Task
	err := db.QueryRow(
		"SELECT id, title, description, done FROM tasks WHERE id = ?",
		idStr,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Done)
	if err == sql.ErrNoRows {
		log.WithError(err).Error("Task not found")
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// Обновить задачу
func updateHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/tasks/")
	var updated Task
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		log.WithError(err).Error("Invalid JSON")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM tasks WHERE id = ?)", idStr).Scan(&exists)
	if err != nil || !exists {
		log.Warn("Task not found for update")
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	var t Task
	db.QueryRow("SELECT id, title, description, done FROM tasks WHERE id = ?", idStr).Scan(&t.ID, &t.Title, &t.Description, &t.Done)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// Удалить задачу
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/tasks/")
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM tasks WHERE id = ?)", idStr).Scan(&exists)
	if err != nil || !exists {
		log.WithError(err).Error("Task not found")
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	_, err = db.Exec("DELETE FROM tasks WHERE id = ?", idStr)
	if err != nil {
		log.WithError(err).Error("Ошибка удаления")
		http.Error(w, "Ошибка удаления", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Обработчик для /tasks
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTasksHandler(w, r)
	case http.MethodPost:
		createTaskHandler(w, r)
	default:
		log.Warn("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Обработчик для /tasks/ (с ID)
func taskByIDHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTaskByIDHandler(w, r)
	case http.MethodPut:
		updateHandler(w, r)
	case http.MethodDelete:
		deleteHandler(w, r)
	default:
		log.Warn("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func logMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"ip":     r.RemoteAddr,
		}).Info("Входящий запрос")
		next(w, r)
	}
}

// Главная
func main() {
	port := getEnv("PORT", "8080")
	dbPath := getEnv("DB_PATH", "./tasks.db")
	logLevel := getEnv("LOG_LEVEL", "info")

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	log.WithFields(logrus.Fields{"port": port, "db": dbPath, "level": level}).Info("Сервер инициализирован")

	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Ошибка открытия БД: ", err)
	}
	defer db.Close()

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		done BOOLEAN DEFAULT 0
	);`
	if _, err := db.Exec(createTableSQL); err != nil {
		log.Fatal("Ошибка создания таблицы: ", err)
	}
	http.HandleFunc("/tasks", logMiddleware(tasksHandler))
	http.HandleFunc("/tasks/", logMiddleware(taskByIDHandler))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: nil,
	}

	go func() {
		log.WithField("port", port).Info("Сервер запущен")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.WithField("timeout", "5s").Info("Получен сигнал завершения. Ожидаем завершения запросов")
	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Ошибка при завершении сервера")
	} else {
		log.Info("Сервер завершён корректно")
	}

	if err := db.Close(); err != nil {
		log.WithError(err).Error("Ошибка закрытия БД")
	} else {
		log.Info("БД закрыта.")
	}
	log.Info("Сервер завершён корректно.")
}
