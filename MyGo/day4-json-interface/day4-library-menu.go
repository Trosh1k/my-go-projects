package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Book struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Year        int    `json:"year"`
	IsAvailable bool   `json:"is_available"`
}

type BookCatalog []Book

func (b Book) String() string {
	return fmt.Sprintf("%s (%s, %d г.) — ID: %s", b.Title, b.Author, b.Year, b.ID)
}

func (c BookCatalog) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(c, "", "   ")
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (c BookCatalog) AddBook(b Book) BookCatalog {
	newCatalog := append(c, b)
	return newCatalog
}

func (c BookCatalog) FindBook(id string) (Book, bool) {
	for _, book := range c {
		if book.ID == id {
			return book, true
		}
	}
	return Book{}, false
}

func (c BookCatalog) BorrowBook(id string) BookCatalog {
	for i := 0; i < len(c); i++ {
		if c[i].ID == id {
			if c[i].IsAvailable {
				c[i].IsAvailable = false
			}
			break
		}
	}
	return c
}

func (c BookCatalog) ReturnBook(id string) BookCatalog {
	for i := 0; i < len(c); i++ {
		if c[i].ID == id {
			if !c[i].IsAvailable {
				c[i].IsAvailable = true
			}
			break
		}
	}
	return c
}

func (c BookCatalog) ListBooks() {
	if len(c) == 0 {
		fmt.Println("Каталог пуст!")
		return
	}

	fmt.Println("Список книг:")
	for _, book := range c {
		status := "[Выдана]"
		if book.IsAvailable {
			status = "[Доступна]"
		}
		fmt.Printf("%s %s (%s, %d г.) — ID: %s\n", status, book.Title, book.Author, book.Year, book.ID)
	}
}

func LoadFromFile(filename string) (BookCatalog, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var catalog BookCatalog
	err = json.Unmarshal(data, &catalog)
	if err != nil {
		return nil, err
	}
	return catalog, nil
}

func readString(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func readInt(prompt string) int {
	for {
		fmt.Print(prompt)
		var val int
		_, err := fmt.Scanln(&val)
		if err == nil {
			return val
		}
		fmt.Println("Ошибка: введите целое число.")
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
	}
}

func main() {
	const fileName = "catalog.json"

	catalog := BookCatalog{}

	loaded, err := LoadFromFile(fileName)
	if err != nil {
		fmt.Println("Не удалось загрузить каталог, создаём новый.")
		catalog = BookCatalog{}
	} else {
		catalog = loaded
		fmt.Println("Каталог загружен из файла.")
	}

	for {
		fmt.Println("\n=== Библиотека ===")
		fmt.Println("1. Добавить книгу")
		fmt.Println("2. Показать все книги")
		fmt.Println("3. Найти книгу по ID")
		fmt.Println("4. Взять книгу (выдать)")
		fmt.Println("5. Вернуть книгу")
		fmt.Println("6. Сохранить в файл")
		fmt.Println("7. Загрузить из файла (перезаписать текущий каталог)")
		fmt.Println("8. Выйти")
		fmt.Print("Выберите действие: ")

		choice := readInt("")

		switch choice {
		case 1:
			id := readString("Введите id книги: ")
			title := readString("Введите название книги: ")
			author := readString("Введите автора: ")
			year := readInt("Введите год издания: ")
			newBook := Book{ID: id, Title: title, Author: author, Year: year, IsAvailable: true}
			catalog = catalog.AddBook(newBook)
			fmt.Println("Книга добавлена.")

		case 2:
			catalog.ListBooks()

		case 3:
			id := readString("Введите id книги: ")
			book, found := catalog.FindBook(id)
			if found {
				fmt.Println("Найдена книга:", book)
			} else {
				fmt.Println("Книга не найдена.")
			}

		case 4:
			id := readString("Введите ID книги для выдачи: ")
			catalog = catalog.BorrowBook(id)
			fmt.Println("Операция выполнена (если книга была доступна, она выдана).")

		case 5:
			id := readString("Введите ID книги для возврата: ")
			catalog = catalog.ReturnBook(id)
			fmt.Println("Операция выполнена (если книга была выдана, она возвращена).")

		case 6:
			err := catalog.SaveToFile(fileName)
			if err != nil {
				fmt.Println("Ошибка сохранения:", err)
			} else {
				fmt.Println("Каталог сохранён в", fileName)
			}

		case 7:
			loaded, err := LoadFromFile(fileName)
			if err != nil {
				fmt.Println("Ошибка загрузки:", err)
			} else {
				catalog = loaded
				fmt.Println("Каталог перезагружен из файла.")
			}

		case 8:
			err := catalog.SaveToFile(fileName)
			if err != nil {
				fmt.Println("Ошибка сохранения перед выходом:", err)
			} else {
				fmt.Println("Каталог сохранён. До свидания!")
			}
			return

		default:
			fmt.Println("Неверный выбор, попробуйте снова.")
		}
	}
}
