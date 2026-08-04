package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

var db *sql.DB

// Получить все задачи
func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, description, done FROM tasks")
	if err != nil {
		http.Error(w, "Ошибка БД", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Done)
		if err != nil {
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
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	result, err := db.Exec(
		"INSERT INTO tasks (title, description, done) VALUES (?, ?, ?)",
		newTask.Title, newTask.Description, false,
	)
	if err != nil {
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
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	_, err := db.Exec(
		"UPDATE tasks SET title = ?, description = ?, done = ? WHERE id = ?",
		updated.Title, updated.Description, updated.Done, idStr,
	)
	if err != nil {
		http.Error(w, "Ошибка обновления", http.StatusInternalServerError)
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
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	_, err = db.Exec("DELETE FROM tasks WHERE id = ?", idStr)
	if err != nil {
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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func logMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s", r.Method, r.URL.Path)
		next(w, r)
	}
}

// Главная
func main() {
	var err error
	db, err = sql.Open("sqlite3", "./tasks.db")
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
	http.HandleFunc("/tasks", tasksHandler)
	http.HandleFunc("/tasks/", taskByIDHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: nil,
	}

	go func() {
		log.Println("Сервер запущен на :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Получен сигнал завершения, закрываем сервер...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Сервер завершил работу с ошибкой", err)
	}

	if err := db.Close(); err != nil {
		log.Println("Ошибка закрытия БД:", err)
	} else {
		log.Println("БД закрыта.")
	}
	log.Println("Сервер завершён корректно.")
}
