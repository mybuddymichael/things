package main

import (
	"encoding/json"
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

// Todo represents a Things.app todo item
type Todo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// OperationResult represents the result of a Things.app operation
type OperationResult struct {
	Success bool
	Message string
}

// getTodosFromList retrieves all todos from the specified list in Things.app as structured data
func getTodosFromList(listName string) ([]Todo, error) {
	escapedListName := strings.ReplaceAll(listName, "'", "\\'")
	jxaScript := fmt.Sprintf(`
try {
    var app = Application('Things3');
    var list = app.lists.byName('%s');
    var todos = list.toDos();
    var result = [];
    for (var i = 0; i < todos.length; i++) {
        result.push({
            name: todos[i].name(),
            status: todos[i].status()
        });
    }
    JSON.stringify(result);
} catch (e) {
    'ERROR: List "%s" not found';
}
`, escapedListName, escapedListName)

	output, err := executor.Execute("osascript", "-l", "JavaScript", "-e", jxaScript)
	if err != nil {
		return nil, fmt.Errorf("error running JXA script: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if strings.HasPrefix(outputStr, "ERROR:") {
		return nil, fmt.Errorf("%s", outputStr)
	}

	var todos []Todo
	if err := json.Unmarshal([]byte(outputStr), &todos); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return todos, nil
}

// addTodoToList adds a new todo to the specified list in Things.app
func addTodoToList(listName, text, tags string) (OperationResult, error) {
	escapedListName := strings.ReplaceAll(listName, "'", "\\'")
	escapedText := strings.ReplaceAll(text, "'", "\\'")
	escapedTags := strings.ReplaceAll(tags, "'", "\\'")

	var todoProperties string
	if tags == "" {
		todoProperties = fmt.Sprintf("{name: '%s'}", escapedText)
	} else {
		todoProperties = fmt.Sprintf("{name: '%s', tagNames: '%s'}", escapedText, escapedTags)
	}

	jxaScript := fmt.Sprintf(`
try {
    var app = Application('Things3');
    var list = app.lists.byName('%s');
    var todo = app.ToDo(%s);
    list.toDos.unshift(todo);
    'SUCCESS';
} catch (e) {
    'ERROR: ' + e.message;
}
`, escapedListName, todoProperties)

	output, err := executor.Execute("osascript", "-l", "JavaScript", "-e", jxaScript)
	if err != nil {
		return OperationResult{}, fmt.Errorf("error running JXA script: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if strings.HasPrefix(outputStr, "ERROR:") {
		return OperationResult{
			Success: false,
			Message: outputStr,
		}, nil
	}

	return OperationResult{
		Success: true,
		Message: fmt.Sprintf("To-do added successfully to list \"%s\"!", listName),
	}, nil
}

// deleteTodoFromList deletes a todo by name from a specific list in Things.app
func deleteTodoFromList(listName, todoName string) (OperationResult, error) {
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
        'SUCCESS';
    } else {
        'ERROR: To-do not found in list';
    }
} catch (e) {
    'ERROR: List not found';
}
`, escapedListName, escapedTodoName)

	output, err := executor.Execute("osascript", "-l", "JavaScript", "-e", jxaScript)
	if err != nil {
		return OperationResult{}, fmt.Errorf("error running JXA script: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if strings.HasPrefix(outputStr, "ERROR:") {
		if strings.Contains(outputStr, "not found in list") {
			return OperationResult{
				Success: false,
				Message: fmt.Sprintf("ERROR: To-do \"%s\" not found in list \"%s\"", todoName, listName),
			}, nil
		}
		return OperationResult{
			Success: false,
			Message: fmt.Sprintf("ERROR: List \"%s\" not found", listName),
		}, nil
	}

	return OperationResult{
		Success: true,
		Message: fmt.Sprintf("To-do \"%s\" deleted successfully from list \"%s\"!", todoName, listName),
	}, nil
}

// moveTodoBetweenLists moves a todo from one list to another in Things.app
func moveTodoBetweenLists(fromList, toList, todoName string) (OperationResult, error) {
	escapedFromList := strings.ReplaceAll(fromList, "\"", "\\\"")
	escapedToList := strings.ReplaceAll(toList, "\"", "\\\"")
	escapedTodoName := strings.ReplaceAll(todoName, "\"", "\\\"")

	applescript := fmt.Sprintf(`
try
    tell application "Things3"
        set todoItem to first to do of list "%s" whose name is "%s"
        move todoItem to list "%s"
        return "SUCCESS"
    end tell
on error errMsg
    if errMsg contains "Can't get" then
        return "ERROR: To-do not found"
    else
        return "ERROR: " & errMsg
    end if
end try
`, escapedFromList, escapedTodoName, escapedToList)

	output, err := executor.Execute("osascript", "-e", applescript)
	if err != nil {
		return OperationResult{}, fmt.Errorf("error running AppleScript: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if strings.HasPrefix(outputStr, "ERROR:") {
		if strings.Contains(outputStr, "not found") {
			return OperationResult{
				Success: false,
				Message: fmt.Sprintf("ERROR: To-do \"%s\" not found in list \"%s\"", todoName, fromList),
			}, nil
		}
		return OperationResult{
			Success: false,
			Message: outputStr,
		}, nil
	}

	return OperationResult{
		Success: true,
		Message: fmt.Sprintf("To-do \"%s\" moved successfully from list \"%s\" to list \"%s\"!", todoName, fromList, toList),
	}, nil
}

// renameTodoInList renames a todo by name in a specific list in Things.app
func renameTodoInList(listName, oldName, newName string) (OperationResult, error) {
	escapedListName := strings.ReplaceAll(listName, "'", "\\'")
	escapedOldName := strings.ReplaceAll(oldName, "'", "\\'")
	escapedNewName := strings.ReplaceAll(newName, "'", "\\'")
	jxaScript := fmt.Sprintf(`
try {
    var app = Application('Things3');
    var list = app.lists.byName('%s');
    var todos = list.toDos();
    var todoFound = false;

    for (var i = 0; i < todos.length; i++) {
        if (todos[i].name() === '%s') {
            todos[i].name = '%s';
            todoFound = true;
            break;
        }
    }

    if (todoFound) {
        'SUCCESS';
    } else {
        'ERROR: To-do not found in list';
    }
} catch (e) {
    'ERROR: List not found';
}
`, escapedListName, escapedOldName, escapedNewName)

	output, err := executor.Execute("osascript", "-l", "JavaScript", "-e", jxaScript)
	if err != nil {
		return OperationResult{}, fmt.Errorf("error running JXA script: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if strings.HasPrefix(outputStr, "ERROR:") {
		if strings.Contains(outputStr, "not found in list") {
			return OperationResult{
				Success: false,
				Message: fmt.Sprintf("ERROR: To-do \"%s\" not found in list \"%s\"", oldName, listName),
			}, nil
		}
		return OperationResult{
			Success: false,
			Message: fmt.Sprintf("ERROR: List \"%s\" not found", listName),
		}, nil
	}

	return OperationResult{
		Success: true,
		Message: fmt.Sprintf("To-do \"%s\" renamed to \"%s\" in list \"%s\"!", oldName, newName, listName),
	}, nil
}
