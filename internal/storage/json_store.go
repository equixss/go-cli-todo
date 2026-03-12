package storage

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
	"sort"

	"github.com/equixss/go-cli-todo/internal/models"
)

type Store interface {
	Load() ([]models.Task, error)
	Save([]models.Task) error
	GetPath() string
}

type JSONStore struct {
	path string
}

func NewJSONStore() (*JSONStore, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(usr.HomeDir, ".todo.json")
	return &JSONStore{path: path}, nil
}

func (s *JSONStore) GetPath() string {
	return s.path
}

func (s *JSONStore) Load() ([]models.Task, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Task{}, nil
		}
		return nil, err
	}

	var tasks []models.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}

	s.reindexTasks(&tasks)
	return tasks, nil
}

func (s *JSONStore) Save(tasks []models.Task) error {
	s.reindexTasks(&tasks)

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Пересчитывает ID, чтобы они всегда шли по порядку 1..N
func (s *JSONStore) reindexTasks(tasks *[]models.Task) {
	for i := range *tasks {
		(*tasks)[i].ID = i + 1
	}
	sort.Slice(*tasks, func(i, j int) bool {
		return (*tasks)[i].Created < (*tasks)[j].Created
	})
}
