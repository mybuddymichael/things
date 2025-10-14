package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestFormatTodosForDisplay(t *testing.T) {
	tests := []struct {
		name     string
		todos    []Todo
		expected string
	}{
		{
			name:     "empty list",
			todos:    []Todo{},
			expected: "",
		},
		{
			name: "single open todo",
			todos: []Todo{
				{Name: "Buy groceries", Status: "open"},
			},
			expected: "○ Buy groceries",
		},
		{
			name: "single completed todo",
			todos: []Todo{
				{Name: "Buy groceries", Status: "completed"},
			},
			expected: "✔︎ Buy groceries",
		},
		{
			name: "single canceled todo",
			todos: []Todo{
				{Name: "Buy groceries", Status: "canceled"},
			},
			expected: "✕ Buy groceries",
		},
		{
			name: "multiple todos with mixed statuses",
			todos: []Todo{
				{Name: "Buy groceries", Status: "open"},
				{Name: "Write report", Status: "completed"},
				{Name: "Call dentist", Status: "canceled"},
			},
			expected: "○ Buy groceries\n✔︎ Write report\n✕ Call dentist",
		},
		{
			name: "two open todos",
			todos: []Todo{
				{Name: "Task 1", Status: "open"},
				{Name: "Task 2", Status: "open"},
			},
			expected: "○ Task 1\n○ Task 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTodosForDisplay(tt.todos)
			if result != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestGetStatusSymbol(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"open", "○ "},
		{"completed", "✔︎ "},
		{"canceled", "✕ "},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run("status_"+tt.status, func(t *testing.T) {
			result := getStatusSymbol(tt.status)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatTodoAsJSONL(t *testing.T) {
	creationDate := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	dueDate := time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		todo     Todo
		validate func(*testing.T, string)
	}{
		{
			name: "simple todo",
			todo: Todo{
				Name:   "Buy groceries",
				Status: "open",
			},
			validate: func(t *testing.T, jsonStr string) {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				if result["name"] != "Buy groceries" {
					t.Errorf("expected name 'Buy groceries', got %v", result["name"])
				}
				if result["status"] != "open" {
					t.Errorf("expected status 'open', got %v", result["status"])
				}
			},
		},
		{
			name: "todo with all fields",
			todo: Todo{
				Name:         "Write report",
				Notes:        "Include quarterly data",
				Status:       "open",
				CreationDate: &creationDate,
				DueDate:      &dueDate,
				TagNames:     []string{"Work", "Important"},
				Area:         "Projects",
				Project:      "Q1 Report",
			},
			validate: func(t *testing.T, jsonStr string) {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
					t.Fatalf("invalid JSON: %v", err)
				}
				if result["name"] != "Write report" {
					t.Errorf("expected name 'Write report', got %v", result["name"])
				}
				if result["notes"] != "Include quarterly data" {
					t.Errorf("expected notes, got %v", result["notes"])
				}
				if result["area"] != "Projects" {
					t.Errorf("expected area 'Projects', got %v", result["area"])
				}
				if result["project"] != "Q1 Report" {
					t.Errorf("expected project 'Q1 Report', got %v", result["project"])
				}
				tags := result["tagNames"].([]interface{})
				if len(tags) != 2 {
					t.Errorf("expected 2 tags, got %d", len(tags))
				}
			},
		},
		{
			name: "todo with empty optional fields",
			todo: Todo{
				Name:   "Simple task",
				Status: "completed",
			},
			validate: func(t *testing.T, jsonStr string) {
				// Should not contain optional fields
				if strings.Contains(jsonStr, "notes") {
					t.Error("should not contain 'notes' field")
				}
				if strings.Contains(jsonStr, "tagNames") {
					t.Error("should not contain 'tagNames' field")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatTodoAsJSONL(tt.todo)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.validate(t, result)
		})
	}
}

func TestFormatOperationResult(t *testing.T) {
	tests := []struct {
		name     string
		result   OperationResult
		expected string
	}{
		{
			name: "success message",
			result: OperationResult{
				Success: true,
				Message: "Todo added successfully",
			},
			expected: "Todo added successfully",
		},
		{
			name: "error message",
			result: OperationResult{
				Success: false,
				Message: "ERROR: List not found",
			},
			expected: "ERROR: List not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatOperationResult(tt.result)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
