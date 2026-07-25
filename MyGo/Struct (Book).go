package main

import (
	"fmt"
)

type Book struct {
	ID          string
	Title       string
	Author      string
	Year        int
	IsAvailable bool
}

type BookCatalog []Book

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
			if c[i].IsAvailable { // если доступна
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
			if !c[i].IsAvailable { // если выдана
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

func main() {
	catalog := BookCatalog{} // пустой каталог
	catalog = catalog.AddBook(Book{"b001", "Война и мир", "Толстой", 1869, true})

	book, found := catalog.FindBook("b002")
	if found {
		fmt.Printf("Найдена книга: %s, автор: %s\n", book.Title, book.Author)
	} else {
		fmt.Println("Книга не найдена")
	}

	catalog = catalog.BorrowBook("b002")

	catalog.ListBooks()
}
