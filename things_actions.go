package main

import (
	"fmt"
	"os/exec"
	"strings"
)

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

	execCmd := exec.Command("osascript", "-l", "JavaScript", "-e", jxaScript)
	output, err := execCmd.Output()
	if err != nil {
		return "", fmt.Errorf("error running JXA script: %v", err)
	}

	return string(output), nil
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

	execCmd := exec.Command("osascript", "-l", "JavaScript", "-e", jxaScript)
	output, err := execCmd.Output()
	if err != nil {
		return "", fmt.Errorf("error running JXA script: %v", err)
	}

	return string(output), nil
}
