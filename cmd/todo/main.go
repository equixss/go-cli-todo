package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/equixss/go-cli-todo/internal/storage"
	"github.com/fatih/color"
)

func main() {
	store, err := storage.NewJSONStore()
	if err != nil {
		color.Red("Ошибка инициализации хранилища: %v", err)
		os.Exit(1)
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("TODO. Введите 'help' для списка команд или 'exit' для выхода.")

	for {
		fmt.Print("\n> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Ошибка чтения:", err)
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		cmd := strings.ToLower(parts[0])
		args := parts[1:]

		if cmd == "exit" || cmd == "quit" {
			fmt.Println("До свидания!")
			break
		}

		HandleCommand(append([]string{cmd}, args...), store)
	}
}
