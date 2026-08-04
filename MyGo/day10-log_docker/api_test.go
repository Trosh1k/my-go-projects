package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestGetTask(t *testing.T) {
	req := httptest.NewRequest("GET", "/tasks", nil)
	w := httptest.NewRecorder()

	getTasksHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидался 200, а получили %d", w.Code)
	}
}

func TestCreateTask(t *testing.T) {
	task := Task{Title: "Тест", Description: "Описание"}
	body, _ := json.Marshal(task)

	req := httptest.NewRequest("POST", "/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	createTaskHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Ожидался 201, а получили %d", w.Code)
	}
}

func TestGetTaskByID_NotFound(t *testing.T) {
	req := httptest.NewRequest("GET", "/tasks/999", nil)
	w := httptest.NewRecorder()

	getTaskByIDHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Ожидался 404, а получили %d", w.Code)
	}
}

func TestUpdateTask_NotFound(t *testing.T) {
	task := Task{Title: "Обновление", Done: true}
	body, _ := json.Marshal(task)

	req := httptest.NewRequest("PUT", "/tasks/999", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	updateHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Ожидался 404, а получили %d", w.Code)
	}
}

func TestDeleteTask_NotFound(t *testing.T) {
	req := httptest.NewRequest("DELETE", "/tasks/999", nil)
	w := httptest.NewRecorder()

	deleteHandler(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Ожидался 404, а получили %d", w.Code)
	}
}
