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

// getTodosFromList retrieves all todos from the specified list in Things.app with status indicators
func getTodosFromList(listName string) (string, error) {
	escapedListName := strings.ReplaceAll(listName, "'", "\\'")
	jxaScript := fmt.Sprintf(`
try {
    var app = Application('Things3');
    var list = app.lists.byName('%s');
    var todos = list.toDos();
    var result = '';
    for (var i = 0; i < todos.length; i++) {
        var status = todos[i].status();
        var symbol = '';
        if (status === 'open') {
            symbol = '○ ';
        } else if (status === 'completed') {
            symbol = '✔︎ ';
        } else if (status === 'canceled') {
            symbol = '✕ ';
        }
        result += symbol + todos[i].name();
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

// deleteTodoFromList deletes a todo by name from a specific list in Things.app
func deleteTodoFromList(listName, todoName string) (string, error) {
	escapedListName := strings.ReplaceAll(listName, "'", "\\'")
	escapedTodoName := strings.ReplaceAll(todoName, "'", "\\'")
	jxaScript := fmt.Sprintf(`
try {
    var app = Application('Things3');
    var list = app.lists.byName('%s');
    var todos = list.toDos();
    var todoFound = false;
    
    for (var i = 0; i < todos.length; i++) {
        if (todos[i].name() === '%s') {
            app.delete(todos[i]);
            todoFound = true;
            break;
        }
    }
    
    if (todoFound) {
        'To-do "%s" deleted successfully from list "%s"!';
    } else {
        'ERROR: To-do "%s" not found in list "%s"';
    }
} catch (e) {
    'ERROR: List "%s" not found';
}
`, escapedListName, escapedTodoName, escapedTodoName, escapedListName, escapedTodoName, escapedListName, escapedListName)

	output, err := executor.Execute("osascript", "-l", "JavaScript", "-e", jxaScript)
	if err != nil {
		return "", fmt.Errorf("error running JXA script: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}
