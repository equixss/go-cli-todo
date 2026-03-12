package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/equixss/go-cli-todo/internal/models"
	"github.com/equixss/go-cli-todo/internal/storage"
	"github.com/fatih/color"
)

func printUsage() {
	fmt.Println(`Команды:
  add <текст> [-p low|medium|high]  Добавить задачу
  list [--done]                     Показать задачи
  edit <N> <текст>                  Редактировать задачу
  done <N>                          Отметить выполненной
  del <N>                           Удалить задачу
  clear                             Очистить все
  exit                              Выход`)
}

func cmdAdd(args []string, store storage.Store) {
	if len(args) == 0 {
		color.Red("Ошибка: не указан текст задачи.")
		return
	}
	var textParts []string
	priorityStr := models.PriorityLow.String()
	for i := 0; i < len(args); i++ {
		if args[i] == "-p" && i+1 < len(args) {
			priorityStr = args[i+1]
			i++
		} else {
			textParts = append(textParts, args[i])
		}
	}
	if len(textParts) == 0 {
		color.Red("Ошибка: не указан текст задачи.")
		return
	}
	text := strings.Join(textParts, " ")
	tasks, err := store.Load()
	if err != nil {
		color.Red("Не удалось загрузить задачи:", err)
		return
	}
	priority, err := models.ParsePriority(priorityStr)
	if err != nil {
		color.Red("Ошибка: недопустимый приоритет '%s'. Допустимые значения: low, medium, high.", priorityStr)
		return
	}
	newTask := models.NewTask(text, priority)
	tasks = append(tasks, newTask)
	if err := store.Save(tasks); err != nil {
		color.Red("Не удалось сохранить задачи:", err)
		return
	}
	color.Green("✓ Добавлена задача: %s", text)
}

func cmdList(args []string, store storage.Store) {
	tasks, err := store.Load()
	if err != nil {
		color.Red("Не удалось загрузить задачи:", err)
		return
	}

	showDone := false
	priorityFilter := ""
	limit := -1

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--done":
			showDone = true
		case "--priority":
			if i+1 < len(args) {
				priorityFilter = args[i+1]
				i++
			} else {
				color.Red("Ошибка: укажите значение приоритета после --priority")
				return
			}
		case "--limit":
			if i+1 < len(args) {
				n, err := strconv.Atoi(args[i+1])
				if err != nil || n <= 0 {
					color.Red("Ошибка: --limit должен быть положительным числом")
					return
				}
				limit = n
				i++
			} else {
				color.Red("Ошибка: укажите число после --limit")
				return
			}
		}
	}

	filtered := tasks.Filter(showDone)
	filtered = filtered.SortByPriority()

	if priorityFilter != "" {
		priority, err := models.ParsePriority(priorityFilter)
		if err != nil {
			color.Red("Ошибка: недопустимый приоритет '%s'. Допустимые: low, medium, high.", priorityFilter)
			return
		}
		var prioritized []models.Task
		for _, t := range filtered {
			if t.Priority == priority {
				prioritized = append(prioritized, t)
			}
		}
		filtered = prioritized
	}

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	if len(filtered) == 0 {
		fmt.Println("Список задач пуст. Добавьте первую задачу командой: add \"...\"")
		return
	}

	header := "📋 TODO‑лист"
	if showDone || priorityFilter != "" || limit > 0 {
		var filters []string
		if showDone {
			filters = append(filters, "выполненные")
		}
		if priorityFilter != "" {
			filters = append(filters, fmt.Sprintf("приоритет: %s", priorityFilter))
		}
		if limit > 0 {
			filters = append(filters, fmt.Sprintf("лимит: %d", limit))
		}
		header = fmt.Sprintf("📋 TODO‑лист (%s)", strings.Join(filters, ", "))
	}
	color.Cyan(header + ":")
	for _, t := range filtered {
		statusIcon := "○"
		statusColor := color.New(color.FgYellow)

		if t.Done {
			statusIcon = "✔"
			statusColor = color.New(color.FgGreen, color.Faint)
		}

		priorityBadge := fmt.Sprintf("[%s] ", t.Priority.String())

		idStr := color.New(color.FgCyan, color.Bold).Sprintf("%d", t.ID)
		timeStr := color.New(color.FgHiBlack).Sprintf("[%s]", t.TimeAgo())

		statusColored := statusColor.Sprint(statusIcon)

		fmt.Printf("%s. %s%s %s %s\n",
			idStr,
			priorityBadge,
			statusColored,
			t.Text,
			timeStr,
		)
	}
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
	force := false
	var taskArgs []string
	for i := 0; i < len(args); i++ {
		if args[i] == "-f" {
			force = true
		} else {
			taskArgs = append(taskArgs, args[i])
		}
	}

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
	if !force {
		fmt.Printf("Вы уверены, что хотите удалить задачу #%d: \"%s\"? (y/N): ", index+1, removed.Text)
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
	}
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
		fmt.Println("Пример: edit 1 \"Новый текст задачи\"")
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
		color.Red("Не удалось сохранить задачи:", err)
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
	if err := store.Clear(); err != nil && !os.IsNotExist(err) {
		color.Red("Не удалось удалить файл:", err)
		return
	}
	color.Red("🗑 Все задачи удалены.")
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
