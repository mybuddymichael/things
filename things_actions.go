package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
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

// JXA code snippet for building a todo item object
// This is the common logic extracted to avoid duplication
const jxaTodoObjectBuilder = `
        var item = {
            name: todo.name(),
            status: todo.status()
        };

        // Add optional string properties
        if (todo.notes()) item.notes = todo.notes();

        // Add date properties (convert to ISO 8601 strings)
        if (todo.creationDate()) item.creationDate = todo.creationDate().toISOString();
        if (todo.modificationDate()) item.modificationDate = todo.modificationDate().toISOString();
        if (todo.dueDate()) item.dueDate = todo.dueDate().toISOString();
        if (completionDate) item.completionDate = completionDate.toISOString();
        if (todo.cancellationDate()) item.cancellationDate = todo.cancellationDate().toISOString();

        // Add tag names (convert string to array if needed)
        var tags = todo.tagNames();
        if (tags) {
            if (typeof tags === 'string') {
                item.tagNames = tags.split(',').map(function(t) { return t.trim(); }).filter(function(t) { return t.length > 0; });
            } else if (tags.length > 0) {
                item.tagNames = tags;
            }
        }

        // Add parent references
        if (todo.area && todo.area()) item.area = todo.area().name();
        if (todo.project && todo.project()) item.project = todo.project().name();

        result.push(item);`

// Todo represents a Things.app todo item with all available properties
type Todo struct {
	// Basic properties
	Name   string `json:"name"`
	Notes  string `json:"notes,omitempty"`
	Status string `json:"status"` // "open", "completed", "canceled"

	// Date properties
	CreationDate     *time.Time `json:"creationDate,omitempty"`
	ModificationDate *time.Time `json:"modificationDate,omitempty"`
	DueDate          *time.Time `json:"dueDate,omitempty"`
	CompletionDate   *time.Time `json:"completionDate,omitempty"`
	CancellationDate *time.Time `json:"cancellationDate,omitempty"`

	// Tags
	TagNames []string `json:"tagNames,omitempty"`

	// Parent references
	Area    string `json:"area,omitempty"`
	Project string `json:"project,omitempty"`
}

// OperationResult represents the result of a Things.app operation
type OperationResult struct {
	Success bool
	Message string
}

// getTodosFromListWithFilter retrieves todos from a list, optionally filtered by completion date
// If filterDateISO is empty, all todos are returned; otherwise, only todos completed after the filter date
func getTodosFromListWithFilter(listName, filterDateISO string) ([]Todo, error) {
	escapedListName := strings.ReplaceAll(listName, "'", "\\'")

	var filterSetup, filterCheck string
	if filterDateISO != "" {
		filterSetup = fmt.Sprintf("var filterDate = new Date('%s');", filterDateISO)
		filterCheck = `
        // Skip if no completion date or before filter date
        if (!completionDate || completionDate < filterDate) {
            continue;
        }`
	}

	jxaScript := fmt.Sprintf(`
try {
    var app = Application('Things3');
    var list = app.lists.byName('%s');
    var todos = list.toDos();
    var result = [];
    %s

    for (var i = 0; i < todos.length; i++) {
        var todo = todos[i];
        var completionDate = todo.completionDate();
%s
%s
    }
    JSON.stringify(result);
} catch (e) {
    'ERROR: List "%s" not found';
}
`, escapedListName, filterSetup, filterCheck, jxaTodoObjectBuilder, escapedListName)

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

// getTodosFromList retrieves all todos from the specified list in Things.app as structured data
func getTodosFromList(listName string) ([]Todo, error) {
	return getTodosFromListWithFilter(listName, "")
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

// logCompletedNow tells Things.app to move completed todos to the Logbook
func logCompletedNow() error {
	jxaScript := `
try {
    var app = Application('Things3');
    app.logCompletedNow();
    'SUCCESS';
} catch (e) {
    'ERROR: ' + e.message;
}
`
	output, err := executor.Execute("osascript", "-l", "JavaScript", "-e", jxaScript)
	if err != nil {
		return fmt.Errorf("error running JXA script: %v", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if strings.HasPrefix(outputStr, "ERROR:") {
		return fmt.Errorf("%s", outputStr)
	}

	return nil
}

// calculateStartDate returns the start date based on the filter
func calculateStartDate(filter string) time.Time {
	now := time.Now()
	switch filter {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "this week":
		// Go back to most recent Sunday at midnight
		daysBack := int(now.Weekday()) // Sunday = 0, Monday = 1, etc.
		sunday := now.AddDate(0, 0, -daysBack)
		return time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 0, 0, 0, 0, sunday.Location())
	case "this month":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	default:
		return time.Time{} // Zero time
	}
}

// parseDateFilter parses a date filter string and returns the start time and whether it represents a single day
// Returns: (startTime, isSingleDay, error)
// - For keywords like "today", "this week", "this month": returns (start of period, false, nil)
// - For YYYY-MM-DD dates: returns (midnight of that day, true, nil)
func parseDateFilter(filter string) (time.Time, bool, error) {
	// Check if it's a keyword
	if filter == "today" || filter == "this week" || filter == "this month" {
		return calculateStartDate(filter), false, nil
	}

	// Try parsing as YYYY-MM-DD
	t, err := time.Parse("2006-01-02", filter)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("invalid date format: %s", filter)
	}

	// Set to midnight in local timezone
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	return startOfDay, true, nil
}

// getCompletedTodos retrieves completed todos from the Logbook filtered by date
func getCompletedTodos(dateFilter string) ([]Todo, error) {
	// First, ensure all completed todos are moved to the Logbook
	if err := logCompletedNow(); err != nil {
		return nil, err
	}

	startDate, isSingleDay, err := parseDateFilter(dateFilter)
	if err != nil {
		return nil, err
	}

	startDateISO := startDate.Format(time.RFC3339)
	todos, err := getTodosFromListWithFilter("Logbook", startDateISO)
	if err != nil {
		return nil, err
	}

	// If filtering for a single day, only include todos completed within that specific day
	if isSingleDay {
		endOfDay := startDate.AddDate(0, 0, 1) // Midnight of next day in local time
		var filtered []Todo
		for _, todo := range todos {
			if todo.CompletionDate != nil {
				// Convert completion date to local timezone for comparison
				completionLocal := todo.CompletionDate.In(time.Local)
				// Include if completion is on or after startDate AND before endOfDay
				if !completionLocal.Before(startDate) && completionLocal.Before(endOfDay) {
					filtered = append(filtered, todo)
				}
			}
		}
		return filtered, nil
	}

	return todos, nil
}

// getCompletedTodosFiltered retrieves completed todos with optional area/project filters
func getCompletedTodosFiltered(dateFilter, areaFilter, projectFilter string) ([]Todo, error) {
	todos, err := getCompletedTodos(dateFilter)
	if err != nil {
		return nil, err
	}

	// If no filters, return all
	if areaFilter == "" && projectFilter == "" {
		return todos, nil
	}

	var filtered []Todo
	for _, todo := range todos {
		// Apply area filter if specified
		if areaFilter != "" && todo.Area != areaFilter {
			continue
		}

		// Apply project filter if specified
		if projectFilter != "" && todo.Project != projectFilter {
			continue
		}

		filtered = append(filtered, todo)
	}
	return filtered, nil
}
