package models

import (
	"testing"
	"time"
)

func TestTimeAgo(t *testing.T) {
	now := time.Now().Unix()

	task := Task{Text: "Task 1", Done: false}

	tests := []struct {
		name     string
		offset   int64
		expected string
	}{
		{"just now", 0, "только что"},
		{"30 sec ago", 30, "только что"},
		{"59 sec ago", 59, "только что"},
		{"1 min ago", 60, "1 мин. назад"},
		{"5 min ago", 300, "5 мин. назад"},
		{"1 hour ago", 3600, "1 ч. назад"},
		{"2 hours ago", 7200, "2 ч. назад"},
		{"1 day ago", 86400, "1 дн. назад"},
		{"5 days ago", 432000, "5 дн. назад"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем время в прошлом: now - offset
			timestamp := now - tt.offset
			task.CreatedAt = timestamp
			got := task.TimeAgo()
			if got != tt.expected {
				t.Errorf("TimeAgo(%d) = %v, want %v", timestamp, got, tt.expected)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	var tasks TaskSlice = []Task{
		{Text: "Task 1", Done: false},
		{Text: "Task 2", Done: true},
		{Text: "Task 3", Done: false},
		{Text: "Task 4", Done: true},
	}

	t.Run("ShowPending", func(t *testing.T) {
		result := tasks.Filter(false)
		if len(result) != 2 {
			t.Errorf("Expected 2 pending tasks, got %d", len(result))
		}
		for _, task := range result {
			if task.Done {
				t.Error("Filter returned a done task when asking for pending")
			}
		}
	})

	t.Run("ShowDone", func(t *testing.T) {
		result := tasks.Filter(true)
		if len(result) != 2 {
			t.Errorf("Expected 2 done tasks, got %d", len(result))
		}
		for _, task := range result {
			if !task.Done {
				t.Error("Filter returned a pending task when asking for done")
			}
		}
	})
}

func TestSortByPriority(t *testing.T) {
	var tasks TaskSlice = []Task{
		{Text: "Low", Priority: PriorityLow},
		{Text: "High", Priority: PriorityHigh},
		{Text: "None", Priority: 32},
		{Text: "Medium", Priority: PriorityMedium},
		{Text: "High2", Priority: PriorityHigh},
	}

	sorted := tasks.SortByPriority()

	expectedOrder := []Priority{Priority(32), PriorityHigh, PriorityHigh, PriorityMedium, PriorityLow}

	if len(sorted) != len(expectedOrder) {
		t.Fatalf("Sorted slice length mismatch: got %d, want %d", len(sorted), len(expectedOrder))
	}

	for i, expPriority := range expectedOrder {
		if sorted[i].Priority != expPriority {
			t.Errorf("At index %d: got priority %q, want %q", i, sorted[i].Priority, expPriority)
		}
	}
}

func TestNewTask(t *testing.T) {
	task := NewTask("Buy milk", PriorityHigh)

	if task.Text != "Buy milk" {
		t.Errorf("Text mismatch: got %q", task.Text)
	}
	if task.Done != false {
		t.Error("New task should not be done")
	}
	if task.Priority != PriorityHigh {
		t.Errorf("Priority mismatch: got %q", task.Priority)
	}
	diff := time.Now().Unix() - task.CreatedAt
	if diff < 0 || diff > 1 {
		t.Errorf("Created timestamp is suspicious: diff=%d", diff)
	}
}

func TestParsePriority(t *testing.T) {
	cases := []struct {
		in   string
		want Priority
	}{
		{"low", PriorityLow},
		{"LOW", PriorityLow},
		{"medium", PriorityMedium},
		{"HIGH", PriorityHigh},
		{" unknown ", Priority(0)},
	}

	for _, c := range cases {
		got, err := ParsePriority(c.in)
		if c.in == " unknown " {
			if err == nil {
				t.Fatalf("ParsePriority(%q) expected error, got nil", c.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("ParsePriority(%q) error: %v", c.in, err)
		}
		if got != c.want {
			t.Fatalf("ParsePriority(%q) = %v; want %v", c.in, got, c.want)
		}
	}
}
