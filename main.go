package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

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
`, listName, listName)

					execCmd := exec.Command("osascript", "-l", "JavaScript", "-e", jxaScript)
					output, err := execCmd.Output()
					if err != nil {
						log.Fatalf("Error running JXA script: %v", err)
					}

					outputStr := string(output)
					if len(outputStr) > 0 && outputStr[:6] == "ERROR:" {
						fmt.Print(outputStr)
						fmt.Println("Use `things list` to see available lists.")
						os.Exit(1)
					}

					fmt.Print(outputStr)
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
`, listName, text, listName)

					execCmd := exec.Command("osascript", "-l", "JavaScript", "-e", jxaScript)
					output, err := execCmd.Output()
					if err != nil {
						log.Fatalf("Error running JXA script: %v", err)
					}

					outputStr := string(output)
					if len(outputStr) > 0 && outputStr[:6] == "ERROR:" {
						fmt.Print(outputStr)
						os.Exit(1)
					}

					fmt.Print(outputStr)
					return nil
				},
			},
		},
	}

	cmd.Run(context.Background(), os.Args)
}
