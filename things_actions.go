package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// CommandExecutor interface allows mocking exec.Command in tests
type CommandExecutor interface {
	Execute(name string, args ...string) ([]byte, error)
}

// DefaultExecutor implements CommandExecutor using real exec.Command
type DefaultExecutor struct{}

func (e *DefaultExecutor) Execute(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

// Global executor - can be replaced in tests
var executor CommandExecutor = &DefaultExecutor{}

// getTodosFromList retrieves all todos from the specified list in Things.app
func getTodosFromList(listName string) (string, error) {
	escapedListName := strings.ReplaceAll(listName, "'", "\\'")
	jxaScript := fmt.Sprintf(`
try {
    var app = Application('Things3');
    var list = app.lists.byName('%s');
    var todos = list.toDos();
    var result = '';
    for (var i = 0; i < todos.length; i++) {
        result += todos[i].name();
        if (i < todos.length - 1) {
            result += '\n';
        }
    }
    result;
} catch (e) {
    'ERROR: List "%s" not found';
}
`, escapedListName, escapedListName)

	output, err := executor.Execute("osascript", "-l", "JavaScript", "-e", jxaScript)
	if err != nil {
		return "", fmt.Errorf("error running JXA script: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// addTodoToList adds a new todo to the specified list in Things.app
func addTodoToList(listName, text string) (string, error) {
	escapedListName := strings.ReplaceAll(listName, "'", "\\'")
	escapedText := strings.ReplaceAll(text, "'", "\\'")
	jxaScript := fmt.Sprintf(`
try {
    var app = Application('Things3');
    var list = app.lists.byName('%s');
    var todo = app.ToDo({name: '%s'});
    list.toDos.unshift(todo);
    'To-do added successfully to list "%s"!';
} catch (e) {
    'ERROR: ' + e.message;
}
`, escapedListName, escapedText, escapedListName)

	output, err := executor.Execute("osascript", "-l", "JavaScript", "-e", jxaScript)
	if err != nil {
		return "", fmt.Errorf("error running JXA script: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// deleteTodoByName deletes a todo by name from Things.app
func deleteTodoByName(todoName string) (string, error) {
	escapedTodoName := strings.ReplaceAll(todoName, "'", "\\'")
	jxaScript := fmt.Sprintf(`
try {
    var app = Application('Things3');
    var todo = app.toDos.byName('%s');
    app.delete(todo);
    'To-do "%s" deleted successfully!';
} catch (e) {
    'ERROR: To-do "%s" not found';
}
`, escapedTodoName, escapedTodoName, escapedTodoName)

	output, err := executor.Execute("osascript", "-l", "JavaScript", "-e", jxaScript)
	if err != nil {
		return "", fmt.Errorf("error running JXA script: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}
