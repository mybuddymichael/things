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
	var text string

	cmd := &cli.Command{
		Name:    "things",
		Version: version,
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
						fmt.Print(output + "\nUse `things list` to see available lists.")
						os.Exit(1)
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
						Name:        "text",
						Aliases:     []string{"t"},
						Usage:       "the `to-do text` to add",
						Required:    true,
						Destination: &text,
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					output, err := addTodoToList(listName, text)
					if err != nil {
						return err
					}
					if strings.HasPrefix(output, "ERROR:") {
						fmt.Print(output)
						os.Exit(1)
					}
					fmt.Print(output)
					return nil
				},
			},
		},
	}

	cmd.Run(context.Background(), os.Args)
}
