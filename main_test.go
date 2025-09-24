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
	originalExecutor := executor
	originalOsExiter := cli.OsExiter
	originalStderr := os.Stderr

	executor = &MockExecutor{
		output: []byte(output),
		err:    err,
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
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					output, err := getTodosFromList(listName)
					if err != nil {
						return err
					}
					if strings.HasPrefix(output, "ERROR:") {
						return cli.Exit(output+"\nUse `things list` to see available lists.", 1)
					}
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
					output, err := addTodoToList(listName, todoName, "")
					if err != nil {
						return err
					}
					if strings.HasPrefix(output, "ERROR:") {
						return cli.Exit(output, 1)
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
					output, err := deleteTodoFromList(listName, todoName)
					if err != nil {
						return err
					}
					if strings.HasPrefix(output, "ERROR:") {
						return cli.Exit(output, 1)
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
					output, err := moveTodoBetweenLists(fromList, toList, todoName)
					if err != nil {
						return err
					}
					if strings.HasPrefix(output, "ERROR:") {
						return cli.Exit(output, 1)
					}
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
	cleanup := setupMockExecutorIntegration("Buy groceries\nWrite report", nil)
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
		name string
		args []string
	}{
		{"show alias", []string{"things", "s", "--list", "Work"}},
		{"add alias", []string{"things", "a", "--name", "Test"}},
		{"delete alias", []string{"things", "d", "--list", "Inbox", "--name", "Test"}},
		{"move alias", []string{"things", "m", "--from", "Inbox", "--to", "Work", "--name", "Test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupMockExecutorIntegration("success", nil)
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
