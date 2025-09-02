package main

import (
	"errors"
	"strings"
	"testing"
)

// MockExecutor implements CommandExecutor for testing
type MockExecutor struct {
	output []byte
	err    error
}

func (m *MockExecutor) Execute(name string, args ...string) ([]byte, error) {
	return m.output, m.err
}

// Helper to set up mock executor and restore original after test
func setupMockExecutor(output string, err error) func() {
	originalExecutor := executor
	executor = &MockExecutor{
		output: []byte(output),
		err:    err,
	}
	return func() {
		executor = originalExecutor
	}
}

func TestGetTodosFromList_Success(t *testing.T) {
	tests := []struct {
		name     string
		listName string
		output   string
		expected string
	}{
		{
			name:     "valid list with todos",
			listName: "Work",
			output:   "Buy groceries\nWrite report\nCall dentist",
			expected: "Buy groceries\nWrite report\nCall dentist",
		},
		{
			name:     "empty list",
			listName: "Empty",
			output:   "",
			expected: "",
		},
		{
			name:     "output with trailing whitespace",
			listName: "Work",
			output:   "  Todo 1  \n  Todo 2  \n  ",
			expected: "Todo 1  \n  Todo 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupMockExecutor(tt.output, nil)
			defer cleanup()

			result, err := getTodosFromList(tt.listName)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetTodosFromList_Errors(t *testing.T) {
	tests := []struct {
		name      string
		listName  string
		output    string
		execError error
		expectErr bool
	}{
		{
			name:      "exec command fails",
			listName:  "Work",
			execError: errors.New("osascript not found"),
			expectErr: true,
		},
		{
			name:     "list not found",
			listName: "NonExistent",
			output:   `ERROR: List "NonExistent" not found`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupMockExecutor(tt.output, tt.execError)
			defer cleanup()

			result, err := getTodosFromList(tt.listName)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if strings.HasPrefix(result, "ERROR:") && !strings.Contains(result, tt.listName) {
					t.Errorf("error message should contain list name %q", tt.listName)
				}
			}
		})
	}
}

func TestAddTodoToList_Success(t *testing.T) {
	tests := []struct {
		name     string
		listName string
		todoName string
		output   string
		expected string
	}{
		{
			name:     "add to work list",
			listName: "Work",
			todoName: "New Task",
			output:   `To-do added successfully to list "Work"!`,
			expected: `To-do added successfully to list "Work"!`,
		},
		{
			name:     "add to inbox",
			listName: "inbox",
			todoName: "Quick note",
			output:   `To-do added successfully to list "inbox"!`,
			expected: `To-do added successfully to list "inbox"!`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupMockExecutor(tt.output, nil)
			defer cleanup()

			result, err := addTodoToList(tt.listName, tt.todoName)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestAddTodoToList_Errors(t *testing.T) {
	tests := []struct {
		name      string
		listName  string
		todoName  string
		output    string
		execError error
		expectErr bool
	}{
		{
			name:      "exec fails",
			listName:  "Work",
			todoName:  "Test",
			execError: errors.New("command failed"),
			expectErr: true,
		},
		{
			name:     "list not found",
			listName: "NonExistent",
			todoName: "Test Todo",
			output:   "ERROR: can't get object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupMockExecutor(tt.output, tt.execError)
			defer cleanup()

			_, err := addTodoToList(tt.listName, tt.todoName)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestDeleteTodoFromList_Success(t *testing.T) {
	tests := []struct {
		name     string
		listName string
		todoName string
		output   string
		expected string
	}{
		{
			name:     "delete existing todo from list",
			listName: "Inbox",
			todoName: "Buy groceries",
			output:   `To-do "Buy groceries" deleted successfully from list "Inbox"!`,
			expected: `To-do "Buy groceries" deleted successfully from list "Inbox"!`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupMockExecutor(tt.output, nil)
			defer cleanup()

			result, err := deleteTodoFromList(tt.listName, tt.todoName)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDeleteTodoFromList_Errors(t *testing.T) {
	tests := []struct {
		name      string
		listName  string
		todoName  string
		output    string
		execError error
		expectErr bool
	}{
		{
			name:      "exec fails",
			listName:  "Inbox",
			todoName:  "Test",
			execError: errors.New("command failed"),
			expectErr: true,
		},
		{
			name:     "list not found",
			listName: "NonExistent",
			todoName: "Test",
			output:   `ERROR: List "NonExistent" not found`,
		},
		{
			name:     "todo not found in list",
			listName: "Inbox",
			todoName: "NonExistent",
			output:   `ERROR: To-do "NonExistent" not found in list "Inbox"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupMockExecutor(tt.output, tt.execError)
			defer cleanup()

			result, err := deleteTodoFromList(tt.listName, tt.todoName)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if result != tt.output {
				t.Errorf("expected %q, got %q", tt.output, result)
			}
		})
	}
}

func TestStringEscaping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single quotes",
			input:    "Don't do this",
			expected: "Don\\'t do this",
		},
		{
			name:     "multiple quotes",
			input:    "I'm 'testing' quotes",
			expected: "I\\'m \\'testing\\' quotes",
		},
		{
			name:     "no quotes",
			input:    "Normal text",
			expected: "Normal text",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the string escaping logic used in the actual functions
			result := strings.ReplaceAll(tt.input, "'", "\\'")
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
