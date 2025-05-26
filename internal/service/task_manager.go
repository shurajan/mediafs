package service

import (
	"errors"
	"sync"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID     string     `json:"id"`
	Status TaskStatus `json:"status"`
	Result string     `json:"result,omitempty"`
	Error  string     `json:"error,omitempty"`
}

type TaskManager struct {
	mu    sync.Mutex
	tasks map[string]*Task
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make(map[string]*Task),
	}
}

func (tm *TaskManager) AddTask(task *Task) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.tasks[task.ID] = task
}

func (tm *TaskManager) UpdateTask(id string, status TaskStatus, result string, err error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if task, ok := tm.tasks[id]; ok {
		task.Status = status
		task.Result = result
		if err != nil {
			task.Error = err.Error()
		}
	}
}

func (tm *TaskManager) GetTask(id string) (*Task, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	task, ok := tm.tasks[id]
	if !ok {
		return nil, errors.New("task not found")
	}
	return task, nil
}
