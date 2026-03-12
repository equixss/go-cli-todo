package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/equixss/go-cli-todo/internal/models"
	"github.com/equixss/go-cli-todo/internal/storage"
	"github.com/fatih/color"
)

func printUsage() {
	fmt.Println(`TODO – консольный список задач

Usage:
  todo add "текст задачи"    – добавить задачу
  todo list                  – показать все задачи
  todo done N                – отметить задачу N выполненной
  todo edit N "текст"        – изменить текст задачи N
  todo del N                 – удалить задачу N
  todo clear                 – удалить все задачи (с запросом подтверждения)`)
}

func cmdAdd(args []string, store storage.Store) {
	if len(args) == 0 {
		color.Yellow("Ошибка: не указан текст задачи.")
		return
	}
	text := strings.Join(args, " ")
	tasks, err := store.Load()
	if err != nil {
		color.Red("Не удалось загрузить задачи:", err)
		return
	}
	newTask := models.Task{Text: text, Done: false, Created: time.Now().Unix(), Priority: models.PriorityMedium}
	tasks = append(tasks, newTask)
	if err := store.Save(tasks); err != nil {
		color.Red("Не удалось сохранить задачи:", err)
		return
	}
	color.Green("✓ Добавлена задача: %s", text)
}

func cmdList(_ []string, store storage.Store) {
	tasks, err := store.Load()
	if err != nil {
		color.Red("Не удалось загрузить задачи:", err)
		return
	}
	if len(tasks) == 0 {
		fmt.Println("Список задач пуст. Добавьте первую задачу командой: todo add \"...\"")
		return
	}

	fmt.Println("📋TODO‑лист:")
	for _, t := range tasks {
		statusIcon := "○"
		statusColor := color.New(color.FgYellow)

		if t.Done {
			statusIcon = "✔"
			statusColor = color.New(color.FgGreen, color.Faint)
		}

		idStr := color.New(color.FgCyan, color.Bold).Sprintf("%d", t.ID)
		timeStr := color.New(color.FgHiBlack).Sprintf("[%s]", t.TimeAgo())

		statusColored := statusColor.Sprint(statusIcon)

		fmt.Printf("%s. %s %s %s\n",
			idStr,
			statusColored,
			t.Text,
			timeStr,
		)
	}
}

func parseIndex(args []string, store storage.Store) (int, error) {
	if len(args) != 1 {
		return -1, fmt.Errorf("укажите номер задачи")
	}
	n, err := strconv.Atoi(args[0])
	if err != nil || n <= 0 {
		return -1, fmt.Errorf("номер должен быть положительным числом")
	}

	tasks, err := store.Load()
	if err != nil {
		return -1, err
	}

	for i, t := range tasks {
		if t.ID == n {
			return i, nil
		}
	}
	return -1, fmt.Errorf("задача #%d не найдена", n)
}

func cmdDone(args []string, store storage.Store) {
	index, err := parseIndex(args, store)
	if err != nil {
		color.Red("%v", err)
		return
	}

	tasks, err := store.Load()
	if err != nil {
		color.Red("Не удалось загрузить задачи:", err)
		return
	}

	if tasks[index].Done {
		fmt.Println("Задача уже отмечена как выполненная.")
		return
	}

	tasks[index].Done = true
	if err := store.Save(tasks); err != nil {
		color.Red("Не удалось сохранить задачи:", err)
		return
	}
	color.Green("✔ Задача %d отмечена выполненной.", index+1)
}

func cmdDel(args []string, store storage.Store) {
	index, err := parseIndex(args, store)
	if err != nil {
		color.Red("%v", err)
		return
	}
	tasks, err := store.Load()
	if err != nil {
		color.Red("Не удалось загрузить задачи:", err)
		return
	}

	removed := tasks[index]
	tasks = append(tasks[:index], tasks[index+1:]...)
	if err := store.Save(tasks); err != nil {
		color.Red("Не удалось сохранить задачи:", err)
		return
	}
	color.Red("✖ Удалена задача %d: %s", removed.ID, removed.Text)
}

func cmdEdit(args []string, store storage.Store) {
	if len(args) < 2 {
		color.Red("Ошибка: укажите номер задачи и новый текст.")
		fmt.Println("Пример: todo edit 1 \"Новый текст задачи\"")
		return
	}

	index, err := parseIndex([]string{args[0]}, store)
	if err != nil {
		color.Red("%v", err)
		return
	}

	tasks, err := store.Load()
	if err != nil {
		color.Red("Не удалось загрузить задачи:", err)
		return
	}

	newText := strings.Join(args[1:], " ")
	oldText := tasks[index].Text
	tasks[index].Text = newText

	if err := store.Save(tasks); err != nil {
		fmt.Println("Не удалось сохранить задачи:", err)
		return
	}
	color.Cyan("✎ Задача %d изменена:\n  Было: %s\n  Стало: %s", index+1, oldText, newText)
}

func cmdClear(store storage.Store) {
	fmt.Print("Вы уверены, что хотите удалить **все** задачи? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	reply, err := reader.ReadString('\n')
	if err != nil {
		color.Red("Ошибка чтения ввода")
		return
	}
	reply = strings.TrimSpace(strings.ToLower(reply))
	if reply != "y" && reply != "yes" {
		fmt.Println("Отмена.")
		return
	}
	if err := os.Remove(store.GetPath()); err != nil && !os.IsNotExist(err) {
		color.Red("Не удалось удалить файл:", err)
		return
	}
	color.Red("🗑 Все задачи удалены.")
}

/*func nowUnix() int64 { return time.Now().Unix() }*/

func handleCommand(args []string, store storage.Store) {
	if len(args) == 0 {
		printUsage()
		return
	}

	cmd := args[0]
	taskArgs := args[1:]

	switch cmd {
	case "add":
		cmdAdd(taskArgs, store)
	case "list", "ls":
		cmdList(taskArgs, store)
	case "done", "check":
		cmdDone(taskArgs, store)
	case "edit", "update":
		cmdEdit(taskArgs, store)
	case "del", "delete", "remove", "rm":
		cmdDel(taskArgs, store)
	case "clear", "reset":
		cmdClear(store)
	case "help", "-h", "--help":
		printUsage()
	default:
		color.Red("Неизвестная команда: %s", cmd)
		printUsage()
	}
}

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

		handleCommand(append([]string{cmd}, args...), store)
	}
}
