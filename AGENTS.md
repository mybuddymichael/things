# things

This is a CLI for interacting with Things.app.

## API

- The Things Applescript API is documented here: https://culturedcode.com/things/support/articles/4562654/
- Review the API page if you are ever unclear about how it works.

## Misc

- Use JXA, not Applescript, for all Applescript code.
- When you need to look up the API, use a subagent and a very specific query. The subagent should do its research using this URL, then provide an answer: https://culturedcode.com/things/support/articles/4562654/ 
- Whenever you're done with changes, run `mise run check`.
