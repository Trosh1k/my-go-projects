package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	urls := []string{
		"https://google.com",
		"https://yandex.ru",
		"https://github.com",
		"https://nonexistent-site.xyz",
	}

	results := make(chan string, len(urls))

	var wg sync.WaitGroup
	wg.Add(len(urls))

	for _, url := range urls {
		go func(u string) {
			defer wg.Done()

			client := http.Client{Timeout: 10 * time.Second}
			resp, err := client.Get(u)
			if err != nil {
				results <- fmt.Sprintf("%s - ОШИБКА: %v", u, err)
				return
			}
			defer resp.Body.Close()
			results <- fmt.Sprintf("%s — OK (%d)", u, resp.StatusCode)
		}(url)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Println("Результаты проверки сайтов:")
	for msg := range results {
		fmt.Println(msg)
	}
}
