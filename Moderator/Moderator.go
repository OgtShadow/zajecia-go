package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var badWords []string

func loadBadWords() ([]string, error) {
	var words []string
	file, err := os.Open("Moderator/badwords.csv")
	if err != nil {
		fmt.Printf("Nie udało się otworzyś pliku: %v\n", err)
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			words = append(words, strings.Split(line, ";")...)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Printf("Błąd podczas odczytywania pliku: %v\n", err)
		return nil, err
	}
	return words, nil
}

func main() {
	var err error
	badWords, err = loadBadWords()
	if err != nil {
		fmt.Printf("Błąd podczas ładowania słów niedozwolonych: %v\n", err)
		return
	}
}
