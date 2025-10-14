package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// formatTodosForDisplay formats a list of todos with status symbols for display
func formatTodosForDisplay(todos []Todo) string {
	var result strings.Builder
	for i, todo := range todos {
		symbol := getStatusSymbol(todo.Status)
		result.WriteString(symbol)
		result.WriteString(todo.Name)
		if i < len(todos)-1 {
			result.WriteString("\n")
		}
	}
	return result.String()
}

// getStatusSymbol returns the display symbol for a todo status
func getStatusSymbol(status string) string {
	switch status {
	case "open":
		return "○ "
	case "completed":
		return "✔︎ "
	case "canceled":
		return "✕ "
	default:
		return ""
	}
}

// formatTodoAsJSONL formats a single todo as a JSONL string
func formatTodoAsJSONL(todo Todo) (string, error) {
	jsonBytes, err := json.Marshal(todo)
	if err != nil {
		return "", fmt.Errorf("error marshaling todo: %v", err)
	}
	return string(jsonBytes), nil
}

// formatOperationResult formats an operation result for display
func formatOperationResult(result OperationResult) string {
	return result.Message
}
