package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
)

var total int

func getPageContent(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func countWordOccurrences(url, word string) (int, error) {
	content, err := getPageContent(url)
	if err != nil {
		return 0, err
	}

	reg := regexp.MustCompile("\\b" + word + "\\b")
	occurrences := reg.FindAllStringIndex(content, -1)

	return len(occurrences), nil
}

func worker(jobs <-chan string, results chan<- string, word string, wg *sync.WaitGroup) {
	defer wg.Done()

	for url := range jobs {
		cnt, err := countWordOccurrences(url, word)
		if err != nil {
			log.Println("error:", err)
			continue
		}

		total += cnt

		results <- fmt.Sprintf("%s: %d occurrences of word '%s'", url, cnt, word)
	}
}

func parseFlags() (string, string, int) {
	filePath := flag.String("f", "", "Path to the file with a list of URLs")
	word := flag.String("w", "", "Word to count occurrences of")
	workers := flag.Int("p", 1, "Number of worker threads")
	flag.Parse()

	return *filePath, *word, *workers
}

func main() {
	filePath, word, workers := parseFlags()

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Error opening file:", err)
	}

	defer file.Close()

	urls := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	jobs := make(chan string, len(urls))
	results := make(chan string, len(urls))

	var wg sync.WaitGroup

	// Запуск воркеров
	for i := 1; i <= workers; i++ {
		wg.Add(1)
		go worker(jobs, results, word, &wg)
	}

	// Отправка задач в канал
	for _, url := range urls {
		jobs <- url
	}

	close(jobs)

	// Получение результатов
	go func() {
		wg.Wait()
		close(results)
	}()

	// Вывод результатов
	for res := range results {
		fmt.Println(res)
	}

	fmt.Println("total:", total)
}
