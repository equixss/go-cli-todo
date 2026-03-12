package models

import (
	"fmt"
	"time"
)

type Task struct {
	ID       int      `json:"id"`
	Text     string   `json:"text"`
	Done     bool     `json:"done"`
	Created  int64    `json:"created"`
	Priority Priority `json:"priority"`
}

func (t *Task) TimeAgo() string {
	diff := time.Now().Unix() - t.Created
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

type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
)

func (p Priority) String() string {
	switch p {
	case PriorityHigh:
		return "HIGH"
	case PriorityMedium:
		return "MEDIUM"
	default:
		return "LOW"
	}
}
