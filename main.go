package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "things",
		Usage: "View your to-dos.",
		Action: func(context.Context, *cli.Command) error {
			// JXA script to fetch today's todos from Things3
			jxaScript := `
var app = Application('Things3');
var today = app.lists.byName('Today');
var todos = today.toDos();
var result = '';
for (var i = 0; i < todos.length; i++) {
    result += todos[i].name();
    if (i < todos.length - 1) {
        result += '\n';
    }
}
result;
`

			cmd := exec.Command("osascript", "-l", "JavaScript", "-e", jxaScript)
			output, err := cmd.Output()
			if err != nil {
				log.Fatalf("Error running JXA script: %v", err)
			}

			fmt.Print(string(output))
			return nil
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
