package main

import (
	"fmt"
)

type Task struct {
	Title       string
	Description string
	Done        bool
}

func main() {
	var tasks []Task

	tasks = AddTask(tasks, "Купить хлеб", "В магазине у дома")
	tasks = AddTask(tasks, "Сделать домашку", "По математике")

	tasks = MarkDone(tasks, "Купить хлеб")

	ListTasks(tasks)
}

func AddTask(tasks []Task, title, desc string) []Task {

	newTask := Task{
		Title:       title,
		Description: desc,
		Done:        false,
	}

	updatedTasks := append(tasks, newTask)

	return updatedTasks
}

func MarkDone(tasks []Task, title string) []Task {
	for i := 0; i < len(tasks); i++ {
		if tasks[i].Title == title {
			tasks[i].Done = true
		}
	}
	return tasks
}

func ListTasks(tasks []Task) {
	if len(tasks) == 0 {
		fmt.Println("Список пуст!")
	}

	fmt.Println("Список задач:")

	for _, task := range tasks {
		status := "[ ]"
		if task.Done {
			status = "[X]"
		}
		fmt.Printf("%s %s: %s\n", status, task.Title, task.Description)
	}
}
