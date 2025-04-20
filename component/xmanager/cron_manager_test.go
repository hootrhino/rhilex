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
	"bytes"
	"log"
	"testing"
	"time"
)

func TestCronManager(t *testing.T) {
	// Create an in-memory buffer to capture log output
	var logBuffer bytes.Buffer
	logger := log.New(&logBuffer, "[CronManager] ", log.LstdFlags)

	// Initialize CronManager with a custom logger
	cm := NewCronManager("../../test")
	cm.SetLogger(logger)

	// Test AddTask
	err := cm.AddTask("task1", "Test Task", "echo 'Hello'", "* * * * *", "Test task description", true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify task was added
	tasks := cm.ListTasks()
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}

	// Test if task logging works
	if !contains(logBuffer.String(), "Task task1 added") {
		t.Fatalf("expected log to contain 'Task task1 added', got: %v", logBuffer.String())
	}

	// Test RemoveTask
	cm.RemoveTask("task1")
	tasks = cm.ListTasks()
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks after removal, got %d", len(tasks))
	}

	// Test if task logging works
	if !contains(logBuffer.String(), "Task task1 removed") {
		t.Fatalf("expected log to contain 'Task task1 removed', got: %v", logBuffer.String())
	}

	// Test Start and Stop
	go cm.Start()
	time.Sleep(1 * time.Second) // Simulate some time for tasks to execute
	cm.Stop()

	// Check if the manager stopped gracefully
	if !contains(logBuffer.String(), "CronManager stopped") {
		t.Fatalf("expected log to contain 'CronManager stopped', got: %v", logBuffer.String())
	}

	// Add a task again to test if it's re-added correctly
	err = cm.AddTask("task2", "Another Task", "echo 'Goodbye'", "* * * * *", "Another test task", true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Simulate a task execution (would be triggered by the cron manager's ticker)
	go cm.executeTask(cm.tasks["task2"])
	time.Sleep(1 * time.Second)

	// Check if execution log is present
	if !contains(logBuffer.String(), "Task task2 executed") {
		t.Fatalf("expected log to contain 'Task task2 executed', got: %v", logBuffer.String())
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
