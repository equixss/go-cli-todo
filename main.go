package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
  todo del N                 – удалить задачу N
  todo clear                 – удалить все задачи (с запросом подтверждения)

Примеры:
  todo add "Купить молоко"
  todo list
  todo done 2
  todo del 1
  todo clear`)
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
	fmt.Println("Добавлена задача:", text)
}

func cmdList(_ []string) {
	tasks, err := loadTasks()
	if err != nil {
		fmt.Println("Не удалось загрузить задачи:", err)
		return
	}
	if len(tasks) == 0 {
		fmt.Println("Список задач пуст.")
		return
	}
	fmt.Println("Ваш TODO‑лист:")
	for i, t := range tasks {
		status := " "
		if t.Done {
			status = "✔"
		}
		fmt.Printf("%2d. [%s] %s\n", i+1, status, t.Text)
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
	fmt.Printf("Задача %d отмечена как выполненная.\n", index+1)
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
	fmt.Printf("Удалена задача %d: %s\n", index+1, removed.Text)
}

func cmdClear(_ []string) {
	fmt.Print("Вы уверены, что хотите удалить **все** задачи? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	reply, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Ошибка чтения ввода")
		return
	}
	if strings.ToLower(reply) != "y" && strings.ToLower(reply) != "yes" {
		fmt.Println("Отмена.")
		return
	}
	if err := os.Remove(storagePathOrEmpty()); err != nil && !os.IsNotExist(err) {
		fmt.Println("Не удалось удалить файл:", err)
		return
	}
	fmt.Println("Все задачи удалены.")
}

func storagePathOrEmpty() string {
	p, _ := storagePath()
	return p
}

func nowUnix() int64 { return time.Now().Unix() }

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		printUsage()
		return
	}

	cmd, cmdArgs := strings.ToLower(args[0]), args[1:]

	switch cmd {
	case "add":
		cmdAdd(cmdArgs)
	case "list":
		cmdList(cmdArgs)
	case "done":
		cmdDone(cmdArgs)
	case "del", "delete", "remove":
		cmdDel(cmdArgs)
	case "clear":
		cmdClear(cmdArgs)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Неизвестная команда: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
