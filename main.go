package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

type Task struct {
	Text    string `json:"text"`
	Done    bool   `json:"done"`
	Created int64  `json:"created"`
}

func storagePath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, ".todo.json"), nil
}

func loadTasks() ([]Task, error) {
	path, err := storagePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Task{}, nil
		}
		return nil, err
	}
	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func saveTasks(tasks []Task) error {
	path, err := storagePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	// Запишем атомарно (чтобы не испортить файл в случае сбоя)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

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

func timeAgo(timestamp int64) string {
	diff := time.Now().Unix() - timestamp
	if diff < 60 {
		return "только что"
	}
	if diff < 3600 {
		return fmt.Sprintf("%d мин. назад", diff/60)
	}
	if diff < 86400 {
		return fmt.Sprintf("%d ч. назад", diff/3600)
	}
	return fmt.Sprintf("%d дн. назад", diff/86400)
}

func cmdAdd(args []string) {
	if len(args) == 0 {
		fmt.Println("Ошибка: не указан текст задачи.")
		return
	}
	text := strings.Join(args, " ")
	tasks, err := loadTasks()
	if err != nil {
		fmt.Println("Не удалось загрузить задачи:", err)
		return
	}
	tasks = append(tasks, Task{Text: text, Done: false, Created: nowUnix()})
	if err := saveTasks(tasks); err != nil {
		fmt.Println("Не удалось сохранить задачи:", err)
		return
	}
	color.Green("✓ Добавлена задача: %s", text)
}

func cmdList(_ []string) {
	tasks, err := loadTasks()
	if err != nil {
		fmt.Println("Не удалось загрузить задачи:", err)
		return
	}
	if len(tasks) == 0 {
		fmt.Println("Список задач пуст. Добавьте первую задачу командой: todo add \"...\"")
		return
	}

	fmt.Println("📋TODO‑лист:")
	for i, t := range tasks {
		idxColor := color.New(color.FgCyan, color.Bold)
		statusIcon := "○"
		statusColor := color.New(color.FgYellow)

		if t.Done {
			statusIcon = "✔"
			statusColor = color.New(color.FgGreen)
		}

		timeStr := color.New(color.FgHiBlack).Sprintf("[%s]", timeAgo(t.Created))

		statusColored := statusColor.Sprint(statusIcon)

		fmt.Printf("%s. %s %s %s\n",
			idxColor.Sprintf("%2d", i+1),
			statusColored,
			t.Text,
			timeStr,
		)
	}
}

func parseTaskIndex(args []string, tasksCount int) int {
	if len(args) != 1 {
		fmt.Println("Ошибка: укажите номер задачи.")
		return -1
	}
	n, err := strconv.Atoi(args[0])
	if err != nil || n <= 0 {
		fmt.Println("Ошибка: номер задачи должен быть положительным целым числом.")
		return -1
	}
	if n > tasksCount {
		fmt.Printf("Ошибка: в списке %d задач(и).\n", tasksCount)
		return -1
	}
	return n - 1
}

func cmdDone(args []string) {
	tasks, err := loadTasks()
	if err != nil {
		fmt.Println("Не удалось загрузить задачи:", err)
		return
	}

	index := parseTaskIndex(args, len(tasks))
	if index == -1 {
		return
	}

	if tasks[index].Done {
		fmt.Println("Задача уже отмечена как выполненная.")
		return
	}

	tasks[index].Done = true
	if err := saveTasks(tasks); err != nil {
		fmt.Println("Не удалось сохранить задачи:", err)
		return
	}
	color.Green("✔ Задача %d отмечена выполненной.", index+1)
}

func cmdDel(args []string) {
	tasks, err := loadTasks()
	if err != nil {
		fmt.Println("Не удалось загрузить задачи:", err)
		return
	}

	index := parseTaskIndex(args, len(tasks))
	if index == -1 {
		return
	}

	removed := tasks[index]
	tasks = append(tasks[:index], tasks[index+1:]...)
	if err := saveTasks(tasks); err != nil {
		fmt.Println("Не удалось сохранить задачи:", err)
		return
	}
	color.Red("✖ Удалена задача %d: %s", index+1, removed.Text)
}

func cmdEdit(args []string) {
	if len(args) < 2 {
		fmt.Println("Ошибка: укажите номер задачи и новый текст.")
		fmt.Println("Пример: todo edit 1 \"Новый текст задачи\"")
		return
	}

	taskNumArgs := []string{args[0]}

	tasks, err := loadTasks()
	if err != nil {
		fmt.Println("Не удалось загрузить задачи:", err)
		return
	}

	idx := parseTaskIndex(taskNumArgs, len(tasks))
	if idx == -1 {
		return
	}

	newText := strings.Join(args[1:], " ")
	oldText := tasks[idx].Text
	tasks[idx].Text = newText

	if err := saveTasks(tasks); err != nil {
		fmt.Println("Не удалось сохранить задачи:", err)
		return
	}
	color.Cyan("✎ Задача %d изменена:\n  Было: %s\n  Стало: %s", idx+1, oldText, newText)
}

func cmdClear(_ []string) {
	fmt.Print("Вы уверены, что хотите удалить **все** задачи? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	reply, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Ошибка чтения ввода")
		return
	}
	reply = strings.TrimSpace(strings.ToLower(reply))
	if reply != "y" && reply != "yes" {
		fmt.Println("Отмена.")
		return
	}
	path, err := storagePath()
	if err != nil {
		fmt.Println("Не удалось определить путь к файлу:", err)
		return
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		fmt.Println("Не удалось удалить файл:", err)
		return
	}
	color.Red("🗑 Все задачи удалены.")
}

func nowUnix() int64 { return time.Now().Unix() }

func main() {
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

		switch cmd {
		case "add":
			cmdAdd(args)
		case "list", "ls":
			cmdList(args)
		case "done", "check":
			cmdDone(args)
		case "edit", "update":
			cmdEdit(args)
		case "del", "delete", "remove", "rm":
			cmdDel(args)
		case "clear", "reset":
			cmdClear(args)
		case "help", "-h", "--help":
			printUsage()
		default:
			color.Red("Неизвестная команда: %s. Введите 'help' для справки.\n", cmd)
		}
	}
}
