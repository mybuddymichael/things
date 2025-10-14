# things

This is a CLI for interacting with Things.app.

## Tech stack

- Go
- urfave/cli/v3 (for command-line interface)

## Things API

- The Things Applescript API is documented here: https://culturedcode.com/things/support/articles/4562654/
- Review the API page if you are ever unclear about how it works.

## Misc

- Use JXA, not Applescript, for all Applescript code.
- Use `mise` for all tasks.
- Whenever you're done with changes, run `mise run check`.
- When you need to look up the API, use a subagent and a very specific query. The subagent should do its research using this URL, then provide an answer: https://culturedcode.com/things/support/articles/4562654/ 
- We track work in Beads instead of Markdown. Run `bd quickstart` to see how.
