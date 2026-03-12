package models

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type TaskSlice []Task
type Task struct {
	ID        int      `json:"id"`
	Text      string   `json:"text"`
	Done      bool     `json:"done"`
	CreatedAt int64    `json:"createdAt"`
	Priority  Priority `json:"priority"`
}

func NewTask(text string, priority Priority) Task {
	return Task{
		Text:      text,
		Done:      false,
		CreatedAt: time.Now().Unix(),
		Priority:  priority,
	}
}

func (t *Task) TimeAgo() string {
	diff := time.Now().Unix() - t.CreatedAt
	if diff < 60 {
		return "только что"
	}
	if diff < 3600 {
		return formatDiff(diff/60, "мин.")
	}
	if diff < 86400 {
		return formatDiff(diff/3600, "ч.")
	}
	return formatDiff(diff/86400, "дн.")
}

func formatDiff(val int64, unit string) string {
	return fmt.Sprintf("%d %s назад", val, unit)
}

func (tasks TaskSlice) Filter(showDone bool) TaskSlice {
	var result []Task
	for _, t := range tasks {
		if showDone {
			if t.Done {
				result = append(result, t)
			}
		} else {
			if !t.Done {
				result = append(result, t)
			}
		}
	}
	return result
}

func (tasks TaskSlice) SortByPriority() TaskSlice {
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Priority > tasks[j].Priority
	})
	return tasks
}

type Priority int8

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "LOW"
	case PriorityMedium:
		return "MEDIUM"
	case PriorityHigh:
		return "HIGH"
	default:
		return fmt.Sprintf("Priority(%d)", int(p))
	}
}

func ParsePriority(priority string) (Priority, error) {
	switch strings.ToUpper(priority) {
	case "LOW":
		return PriorityLow, nil
	case "HIGH":
		return PriorityHigh, nil
	case "MEDIUM":
		return PriorityMedium, nil
	default:
		return 0, fmt.Errorf("неизвестный приоритет %q", priority)
	}
}

func MustParsePriority(s string) Priority {
	p, err := ParsePriority(s)
	if err != nil {
		panic(err)
	}
	return p
}
