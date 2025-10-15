package main

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

// setupMockExecutor sets up a mock executor for testing and disables os.Exit
func setupMockExecutorIntegration(output string, err error) func() {
	return setupMockExecutorIntegrationMulti([]string{output}, []error{err})
}

// setupMockExecutorIntegrationMulti sets up a mock executor with multiple outputs for testing and disables os.Exit
func setupMockExecutorIntegrationMulti(outputs []string, errors []error) func() {
	originalExecutor := executor
	originalOsExiter := cli.OsExiter
	originalStderr := os.Stderr

	byteOutputs := make([][]byte, len(outputs))
	for i, output := range outputs {
		byteOutputs[i] = []byte(output)
	}

	executor = &MockExecutor{
		outputs: byteOutputs,
		errors:  errors,
	}

	// Override OsExiter to prevent actual exit during tests
	cli.OsExiter = func(code int) {
		// Do nothing - just capture that exit was called
	}

	// Redirect stderr to discard CLI error output
	r, w, _ := os.Pipe()
	os.Stderr = w
	go func() {
		// Discard stderr output
		_, _ = io.Copy(io.Discard, r)
	}()

	return func() {
		executor = originalExecutor
		cli.OsExiter = originalOsExiter
		os.Stderr = originalStderr
	}
}

// createTestApp creates the CLI app for testing - same as main.go but testable
func createTestApp() *cli.Command {
	return createTestAppWithWriters(io.Discard, io.Discard)
}

// createTestAppWithWriters creates the CLI app with custom writers for suppressing output
func createTestAppWithWriters(writer, errWriter io.Writer) *cli.Command {
	var listName string
	var todoName string
	var fromList string
	var toList string
	var tags string
	var newName string
	var dateFilter string
	var areaFilter string
	var projectFilter string
	var jsonl bool

	app := &cli.Command{
		Name:    "things",
		Version: "test",
		Usage:   "Interact with Things.app from the command line.",
		Commands: []*cli.Command{
			{
				Name:    "show",
				Usage:   "Show to-dos from a specified list",
				Aliases: []string{"s"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "list",
						Aliases:     []string{"l"},
						Usage:       "show to-dos from the specified `list`",
						Required:    true,
						Destination: &listName,
					},
					&cli.BoolFlag{
						Name:        "jsonl",
						Usage:       "output todos in JSONL format",
						Destination: &jsonl,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					todos, err := getTodosFromList(listName)
					if err != nil {
						if strings.HasPrefix(err.Error(), "ERROR:") {
							return cli.Exit(err.Error()+"\nUse `things list` to see available lists.", 1)
						}
						return err
					}
					_ = todos
					_ = jsonl
					return nil
				},
			},
			{
				Name:    "add",
				Usage:   "Add a new todo to a specified list",
				Aliases: []string{"a"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "list",
						Aliases:     []string{"l"},
						Usage:       "the `list` to add the to-do to",
						Value:       "inbox",
						Destination: &listName,
					},
					&cli.StringFlag{
						Name:        "name",
						Aliases:     []string{"n"},
						Usage:       "the `to-do name` to add",
						Required:    true,
						Destination: &todoName,
					},
					&cli.StringFlag{
						Name:        "tags",
						Aliases:     []string{"t"},
						Usage:       "comma-separated `tags` to add to the to-do (e.g., \"Home, Work\")",
						Destination: &tags,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					result, err := addTodoToList(listName, todoName, tags)
					if err != nil {
						return err
					}
					if !result.Success {
						return cli.Exit(result.Message, 1)
					}
					return nil
				},
			},
			{
				Name:    "delete",
				Usage:   "Delete a todo by name from a specified list",
				Aliases: []string{"d"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "list",
						Aliases:     []string{"l"},
						Usage:       "the `list` to search for the to-do in",
						Required:    true,
						Destination: &listName,
					},
					&cli.StringFlag{
						Name:        "name",
						Aliases:     []string{"n"},
						Usage:       "the `name` of the to-do to delete",
						Required:    true,
						Destination: &todoName,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					result, err := deleteTodoFromList(listName, todoName)
					if err != nil {
						return err
					}
					if !result.Success {
						return cli.Exit(result.Message, 1)
					}
					return nil
				},
			},
			{
				Name:    "move",
				Usage:   "Move a todo from one list to another",
				Aliases: []string{"m"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "from",
						Usage:       "the `list` to move the to-do from",
						Required:    true,
						Destination: &fromList,
					},
					&cli.StringFlag{
						Name:        "to",
						Usage:       "the `list` to move the to-do to",
						Required:    true,
						Destination: &toList,
					},
					&cli.StringFlag{
						Name:        "name",
						Aliases:     []string{"n"},
						Usage:       "the `name` of the to-do to move",
						Required:    true,
						Destination: &todoName,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					result, err := moveTodoBetweenLists(fromList, toList, todoName)
					if err != nil {
						return err
					}
					if !result.Success {
						return cli.Exit(result.Message, 1)
					}
					return nil
				},
			},
			{
				Name:    "rename",
				Usage:   "Rename a todo in a specified list",
				Aliases: []string{"r"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "list",
						Aliases:     []string{"l"},
						Usage:       "the `list` containing the to-do",
						Required:    true,
						Destination: &listName,
					},
					&cli.StringFlag{
						Name:        "name",
						Aliases:     []string{"n"},
						Usage:       "the current `name` of the to-do",
						Required:    true,
						Destination: &todoName,
					},
					&cli.StringFlag{
						Name:        "new-name",
						Usage:       "the `new name` for the to-do",
						Required:    true,
						Destination: &newName,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					result, err := renameTodoInList(listName, todoName, newName)
					if err != nil {
						return err
					}
					if !result.Success {
						return cli.Exit(result.Message, 1)
					}
					return nil
				},
			},
			{
				Name:    "log",
				Usage:   "Show completed to-dos from the Logbook",
				Aliases: []string{"lg"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "date",
						Aliases:     []string{"d"},
						Usage:       "show completed to-dos from `TIMEFRAME` (today, this week, this month)",
						Required:    true,
						Destination: &dateFilter,
					},
					&cli.StringFlag{
						Name:        "area",
						Aliases:     []string{"a"},
						Usage:       "filter by `AREA` name",
						Destination: &areaFilter,
					},
					&cli.StringFlag{
						Name:        "project",
						Aliases:     []string{"p"},
						Usage:       "filter by `PROJECT` name",
						Destination: &projectFilter,
					},
					&cli.BoolFlag{
						Name:        "jsonl",
						Usage:       "output todos in JSONL format",
						Destination: &jsonl,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if dateFilter != "today" && dateFilter != "this week" && dateFilter != "this month" {
						return cli.Exit("ERROR: --date must be one of: today, this week, this month", 1)
					}
					todos, err := getCompletedTodosFiltered(dateFilter, areaFilter, projectFilter)
					if err != nil {
						if strings.HasPrefix(err.Error(), "ERROR:") {
							return cli.Exit(err.Error(), 1)
						}
						return err
					}
					_ = todos
					_ = jsonl
					return nil
				},
			},
		},
	}

	if writer != nil {
		app.Writer = writer
	}
	if errWriter != nil {
		app.ErrWriter = errWriter
	}

	return app
}

func TestShowCommand_RequiredFlag(t *testing.T) {
	cleanup := setupMockExecutorIntegration("", nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "show"})

	// Should fail due to missing required flag
	if err == nil {
		t.Error("expected error for missing required --list flag")
	}
}

func TestShowCommand_Success(t *testing.T) {
	cleanup := setupMockExecutorIntegration(`[{"name":"Buy groceries","status":"open"},{"name":"Write report","status":"open"}]`, nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "show", "--list", "Work"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestShowCommand_ListNotFound(t *testing.T) {
	cleanup := setupMockExecutorIntegration(`ERROR: List "NonExistent" not found`, nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "show", "--list", "NonExistent"})

	// Should return cli.Exit error
	if err == nil {
		t.Error("expected cli.Exit error for non-existent list")
	}

	// Check if it's a cli.Exit error with correct exit code
	if exitErr, ok := err.(cli.ExitCoder); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
		if !strings.Contains(err.Error(), "ERROR:") {
			t.Error("exit error should contain ERROR message")
		}
		if !strings.Contains(err.Error(), "Use `things list`") {
			t.Error("exit error should contain helpful message")
		}
	} else {
		t.Errorf("expected cli.ExitCoder, got %T", err)
	}
}

func TestShowCommand_ExecError(t *testing.T) {
	cleanup := setupMockExecutorIntegration("", errors.New("osascript not found"))
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "show", "--list", "Work"})

	// Should return the exec error, not cli.Exit
	if err == nil {
		t.Error("expected error when exec fails")
	}

	// Should NOT be a cli.Exit error since this is an exec failure
	if _, ok := err.(cli.ExitCoder); ok {
		t.Error("should not be cli.ExitCoder for exec failures")
	}
}

func TestAddCommand_Success(t *testing.T) {
	cleanup := setupMockExecutorIntegration(`To-do added successfully to list "inbox"!`, nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "add", "--name", "Test Todo"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddCommand_Error(t *testing.T) {
	cleanup := setupMockExecutorIntegration("ERROR: List not found", nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "add", "--list", "NonExistent", "--name", "Test"})

	// Should return cli.Exit error
	if err == nil {
		t.Error("expected cli.Exit error for non-existent list")
	}

	if exitErr, ok := err.(cli.ExitCoder); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	} else {
		t.Errorf("expected cli.ExitCoder, got %T", err)
	}
}

func TestDeleteCommand_Success(t *testing.T) {
	cleanup := setupMockExecutorIntegration(`To-do "Test Todo" deleted successfully from list "Inbox"!`, nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "delete", "--list", "Inbox", "--name", "Test Todo"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteCommand_Error(t *testing.T) {
	cleanup := setupMockExecutorIntegration(`ERROR: To-do "NonExistent" not found in list "Inbox"`, nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "delete", "--list", "Inbox", "--name", "NonExistent"})

	// Should return cli.Exit error
	if err == nil {
		t.Error("expected cli.Exit error for non-existent todo")
	}

	if exitErr, ok := err.(cli.ExitCoder); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	} else {
		t.Errorf("expected cli.ExitCoder, got %T", err)
	}
}

func TestCommandAliases(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		mockOutputs []string
	}{
		{"show alias", []string{"things", "s", "--list", "Work"}, []string{`[{"name":"Test","status":"open"}]`}},
		{"add alias", []string{"things", "a", "--name", "Test"}, []string{`To-do added successfully to list "inbox"!`}},
		{"delete alias", []string{"things", "d", "--list", "Inbox", "--name", "Test"}, []string{`To-do "Test" deleted successfully from list "Inbox"!`}},
		{"move alias", []string{"things", "m", "--from", "Inbox", "--to", "Work", "--name", "Test"}, []string{`To-do "Test" moved successfully from list "Inbox" to list "Work"!`}},
		{"rename alias", []string{"things", "r", "--list", "Inbox", "--name", "Old", "--new-name", "New"}, []string{"SUCCESS"}},
		{"log alias", []string{"things", "lg", "--date", "today"}, []string{"SUCCESS", `[{"name":"Completed task","status":"completed"}]`}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := make([]error, len(tt.mockOutputs))
			cleanup := setupMockExecutorIntegrationMulti(tt.mockOutputs, errors)
			defer cleanup()

			app := createTestApp()
			err := app.Run(context.Background(), tt.args)
			if err != nil {
				t.Errorf("alias should work: %v", err)
			}
		})
	}
}

func TestMoveCommand_Success(t *testing.T) {
	cleanup := setupMockExecutorIntegration(`To-do "Test Todo" moved successfully from list "Inbox" to list "Work"!`, nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "move", "--from", "Inbox", "--to", "Work", "--name", "Test Todo"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMoveCommand_Error(t *testing.T) {
	cleanup := setupMockExecutorIntegration(`ERROR: To-do "NonExistent" not found in list "Inbox"`, nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "move", "--from", "Inbox", "--to", "Work", "--name", "NonExistent"})

	// Should return cli.Exit error
	if err == nil {
		t.Error("expected cli.Exit error for non-existent todo")
	}

	if exitErr, ok := err.(cli.ExitCoder); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	} else {
		t.Errorf("expected cli.ExitCoder, got %T", err)
	}
}

func TestMoveCommand_TodayToInbox(t *testing.T) {
	cleanup := setupMockExecutorIntegration(`To-do "Make a small plan for how to help cutter" moved successfully from list "today" to list "inbox"!`, nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "move", "--from", "today", "--to", "inbox", "--name", "Make a small plan for how to help cutter"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddCommand_WithTags(t *testing.T) {
	cleanup := setupMockExecutorIntegration(`To-do added successfully to list "Work"!`, nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "add", "--name", "Test Todo", "--list", "Work", "--tags", "Important, Urgent"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddCommand_WithTagsAlias(t *testing.T) {
	cleanup := setupMockExecutorIntegration(`To-do added successfully to list "inbox"!`, nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "add", "--name", "Test Todo", "-t", "Home"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestFlagValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"show missing list", []string{"things", "show"}},
		{"add missing name", []string{"things", "add"}},
		{"delete missing list", []string{"things", "delete", "--name", "Test"}},
		{"delete missing name", []string{"things", "delete", "--list", "Inbox"}},
		{"move missing from", []string{"things", "move", "--to", "Work", "--name", "Test"}},
		{"move missing to", []string{"things", "move", "--from", "Inbox", "--name", "Test"}},
		{"move missing name", []string{"things", "move", "--from", "Inbox", "--to", "Work"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupMockExecutorIntegration("", nil)
			defer cleanup()

			app := createTestApp()
			err := app.Run(context.Background(), tt.args)

			if err == nil {
				t.Error("expected error for missing required flag")
			}
		})
	}
}

func TestRenameCommand_Success(t *testing.T) {
	cleanup := setupMockExecutorIntegration("SUCCESS", nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "rename", "--list", "Inbox", "--name", "Old Name", "--new-name", "New Name"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRenameCommand_Error(t *testing.T) {
	cleanup := setupMockExecutorIntegration(`ERROR: To-do "NonExistent" not found in list "Inbox"`, nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "rename", "--list", "Inbox", "--name", "NonExistent", "--new-name", "New Name"})

	if err == nil {
		t.Error("expected cli.Exit error for non-existent todo")
	}

	if exitErr, ok := err.(cli.ExitCoder); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
	} else {
		t.Errorf("expected cli.ExitCoder, got %T", err)
	}
}

func TestRenameCommand_MissingFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"missing list", []string{"things", "rename", "--name", "Test", "--new-name", "New"}},
		{"missing name", []string{"things", "rename", "--list", "Inbox", "--new-name", "New"}},
		{"missing new-name", []string{"things", "rename", "--list", "Inbox", "--name", "Test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupMockExecutorIntegration("", nil)
			defer cleanup()

			app := createTestApp()
			err := app.Run(context.Background(), tt.args)

			if err == nil {
				t.Error("expected error for missing required flag")
			}
		})
	}
}

func TestRenameCommand_Alias(t *testing.T) {
	cleanup := setupMockExecutorIntegration("SUCCESS", nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "r", "--list", "Inbox", "--name", "Old", "--new-name", "New"})
	if err != nil {
		t.Errorf("rename alias should work: %v", err)
	}
}

func TestLogCommand_Success(t *testing.T) {
	mockOutput := `[{"name":"Completed task 1","status":"completed"},{"name":"Completed task 2","status":"completed"}]`

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "log today",
			args: []string{"things", "log", "--date", "today"},
		},
		{
			name: "log this week",
			args: []string{"things", "log", "--date", "this week"},
		},
		{
			name: "log this month",
			args: []string{"things", "log", "--date", "this month"},
		},
		{
			name: "log with date alias",
			args: []string{"things", "log", "-d", "today"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock both logCompletedNow() and getTodosFromListWithFilter() calls
			cleanup := setupMockExecutorIntegrationMulti([]string{"SUCCESS", mockOutput}, []error{nil, nil})
			defer cleanup()

			app := createTestApp()
			err := app.Run(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLogCommand_WithFilters(t *testing.T) {
	mockOutput := `[{"name":"Task 1","status":"completed","area":"Work"}]`

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "log with area filter",
			args: []string{"things", "log", "--date", "today", "--area", "Work"},
		},
		{
			name: "log with project filter",
			args: []string{"things", "log", "--date", "today", "--project", "Project A"},
		},
		{
			name: "log with both filters",
			args: []string{"things", "log", "--date", "today", "--area", "Work", "--project", "Project A"},
		},
		{
			name: "log with area alias",
			args: []string{"things", "log", "-d", "today", "-a", "Work"},
		},
		{
			name: "log with project alias",
			args: []string{"things", "log", "-d", "today", "-p", "Project A"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock both logCompletedNow() and getTodosFromListWithFilter() calls
			cleanup := setupMockExecutorIntegrationMulti([]string{"SUCCESS", mockOutput}, []error{nil, nil})
			defer cleanup()

			app := createTestApp()
			err := app.Run(context.Background(), tt.args)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLogCommand_InvalidDateFilter(t *testing.T) {
	cleanup := setupMockExecutorIntegration("", nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "log", "--date", "yesterday"})

	if err == nil {
		t.Error("expected error for invalid date filter")
	}

	if exitErr, ok := err.(cli.ExitCoder); ok {
		if exitErr.ExitCode() != 1 {
			t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
		}
		if !strings.Contains(err.Error(), "ERROR:") {
			t.Error("error should contain ERROR message")
		}
	} else {
		t.Errorf("expected cli.ExitCoder, got %T", err)
	}
}

func TestLogCommand_MissingDateFlag(t *testing.T) {
	cleanup := setupMockExecutorIntegration("", nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "log"})

	if err == nil {
		t.Error("expected error for missing required --date flag")
	}
}

func TestLogCommand_Alias(t *testing.T) {
	mockOutput := `[{"name":"Completed task","status":"completed"}]`
	// Mock both logCompletedNow() and getTodosFromListWithFilter() calls
	cleanup := setupMockExecutorIntegrationMulti([]string{"SUCCESS", mockOutput}, []error{nil, nil})
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "lg", "--date", "today"})
	if err != nil {
		t.Errorf("log alias should work: %v", err)
	}
}

func TestJSONLOutput_Show(t *testing.T) {
	mockOutput := `[{"name":"Task 1","status":"open"},{"name":"Task 2","status":"completed"}]`
	cleanup := setupMockExecutorIntegration(mockOutput, nil)
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "show", "--list", "Work", "--jsonl"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestJSONLOutput_Log(t *testing.T) {
	mockOutput := `[{"name":"Completed task","status":"completed"}]`
	// Mock both logCompletedNow() and getTodosFromListWithFilter() calls
	cleanup := setupMockExecutorIntegrationMulti([]string{"SUCCESS", mockOutput}, []error{nil, nil})
	defer cleanup()

	app := createTestApp()
	err := app.Run(context.Background(), []string{"things", "log", "--date", "today", "--jsonl"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
