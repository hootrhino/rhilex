// Copyright (C) 2025 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package xmanager

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Task represents a scheduled cron task
type Task struct {
	ID          string    // Unique identifier
	Name        string    // Display name
	Command     string    // Command to execute
	Schedule    string    // Cron schedule expression
	CreatedAt   time.Time // Creation timestamp
	LastRun     time.Time // Last execution time
	NextRun     time.Time // Next scheduled execution time
	IsActive    bool      // Whether task is active
	Description string    // Optional description
	LogFile     string    // Optional log file path
}

// CronManager handles cron task operations
type CronManager struct {
	tasks    map[string]*Task // In-memory task storage
	mutex    sync.RWMutex     // For thread safety
	logDir   string           // Directory for log files
	logger   *log.Logger      // Logger for output
	stopChan chan struct{}    // Channel to signal stopping tasks
}

// NewCronManager initializes a new CronManager with a default logger outputting to stdout
func NewCronManager(logDir string) *CronManager {
	return &CronManager{
		tasks:    make(map[string]*Task),
		logDir:   logDir,
		stopChan: make(chan struct{}),
		logger:   log.New(os.Stdout, "[CronManager] ", log.LstdFlags), // Default logger to stdout
	}
}

// SetLogger sets a custom logger for the CronManager
func (cm *CronManager) SetLogger(logger *log.Logger) {
	cm.logger = logger
}

// AddTask adds a new task to the cron manager
func (cm *CronManager) AddTask(id, name, command, schedule, description string, isActive bool) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// Check if task already exists
	if _, exists := cm.tasks[id]; exists {
		return fmt.Errorf("task with ID %s already exists", id)
	}

	// Create the task and add it to the tasks map
	task := &Task{
		ID:          id,
		Name:        name,
		Command:     command,
		Schedule:    schedule,
		CreatedAt:   time.Now(),
		IsActive:    isActive,
		Description: description,
		LogFile:     fmt.Sprintf("%s/%s.log", cm.logDir, id),
	}

	// Set the next scheduled run time (simplified for now)
	task.NextRun = time.Now().Add(time.Minute) // You should use a proper cron parser here
	cm.tasks[id] = task

	cm.logger.Printf("Task %s added", id)
	return nil
}

// RemoveTask removes a task from the cron manager
func (cm *CronManager) RemoveTask(id string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.tasks, id)
	cm.logger.Printf("Task %s removed", id)
}

// Start starts the cron manager and begins executing tasks
func (cm *CronManager) Start() {
	for {
		select {
		case <-cm.stopChan:
			return
		default:
			cm.executeScheduledTasks()
			time.Sleep(1 * time.Minute) // Check tasks every minute
		}
	}
}

// Stop gracefully stops the cron manager
func (cm *CronManager) Stop() {
	close(cm.stopChan)
	cm.logger.Println("CronManager stopped")
}

// executeScheduledTasks executes tasks whose next run time has passed
func (cm *CronManager) executeScheduledTasks() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for _, task := range cm.tasks {
		if task.IsActive && time.Now().After(task.NextRun) {
			// Execute the task
			go cm.executeTask(task)
			// Update task's last run time and schedule the next run
			task.LastRun = time.Now()
			task.NextRun = time.Now().Add(time.Minute) // Update with the next scheduled time
		}
	}
}

// executeTask executes the given task and logs the output
func (cm *CronManager) executeTask(task *Task) {
	// Log task execution
	logFile, err := os.OpenFile(task.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		cm.logger.Printf("Error opening log file for task %s: %v", task.ID, err)
		return
	}
	defer logFile.Close()

	// Execute the task's command (simplified for example)
	fmt.Fprintf(logFile, "Task %s executed at %v\n", task.ID, time.Now())
	// Here, you'd execute the task's command (e.g., using `os/exec` package)
	cm.logger.Printf("Task %s executed successfully", task.ID)
}

// ListTasks returns all the tasks in the cron manager
func (cm *CronManager) ListTasks() []*Task {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	var tasks []*Task
	for _, task := range cm.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}
