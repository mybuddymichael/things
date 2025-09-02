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
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					output, err := getTodosFromList(listName)
					if err != nil {
						return err
					}
					if strings.HasPrefix(output, "ERROR:") {
						return cli.Exit(output+"\nUse `things list` to see available lists.", 1)
					}
					fmt.Print(output)
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
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					output, err := addTodoToList(listName, todoName)
					if err != nil {
						return err
					}
					if strings.HasPrefix(output, "ERROR:") {
						return cli.Exit(output, 1)
					}
					fmt.Print(output)
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
					fmt.Print(output)
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
