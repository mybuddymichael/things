package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
)

var version = "dev"

func main() {
	var listName string
	var todoName string
	var fromList string
	var toList string
	var tags string
	var newName string
	var dateFilter string
	var jsonl bool
	var runs int

	cmd := &cli.Command{
		Name:                  "things",
		Version:               version,
		Usage:                 "Interact with Things.app from the command line.",
		EnableShellCompletion: true,
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

					if jsonl {
						for _, todo := range todos {
							jsonLine, err := formatTodoAsJSONL(todo)
							if err != nil {
								return err
							}
							fmt.Println(jsonLine)
						}
						return nil
					}

					output := formatTodosForDisplay(todos)
					fmt.Println(output)
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
					fmt.Println(formatOperationResult(result))
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
					fmt.Println(formatOperationResult(result))
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
					fmt.Println(formatOperationResult(result))
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
					fmt.Println(formatOperationResult(result))
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
					&cli.BoolFlag{
						Name:        "jsonl",
						Usage:       "output todos in JSONL format",
						Destination: &jsonl,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					// Validate date filter
					if dateFilter != "today" && dateFilter != "this week" && dateFilter != "this month" {
						return cli.Exit("ERROR: --date must be one of: today, this week, this month", 1)
					}

					todos, err := getCompletedTodos(dateFilter)
					if err != nil {
						if strings.HasPrefix(err.Error(), "ERROR:") {
							return cli.Exit(err.Error(), 1)
						}
						return err
					}

					if jsonl {
						for _, todo := range todos {
							jsonLine, err := formatTodoAsJSONL(todo)
							if err != nil {
								return err
							}
							fmt.Println(jsonLine)
						}
						return nil
					}

					output := formatTodosForDisplay(todos)
					fmt.Println(output)
					return nil
				},
			},
			{
				Name:    "experiment",
				Usage:   "Run performance experiments on different fetch approaches",
				Aliases: []string{"exp"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "date",
						Aliases:     []string{"d"},
						Usage:       "test with date filter `TIMEFRAME` (today, this week, this month)",
						Value:       "this week",
						Destination: &dateFilter,
					},
					&cli.IntFlag{
						Name:        "runs",
						Aliases:     []string{"r"},
						Usage:       "number of `runs` per approach",
						Value:       3,
						Destination: &runs,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					// Validate date filter
					if dateFilter != "today" && dateFilter != "this week" && dateFilter != "this month" {
						return cli.Exit("ERROR: --date must be one of: today, this week, this month", 1)
					}

					runAllExperiments(dateFilter, runs)
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
