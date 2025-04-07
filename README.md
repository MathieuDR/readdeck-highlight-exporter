# Readdeck Highlight Exporter

A command-line tool that exports your highlights from Readdeck (a read-it-later service) to your Zettelkasten note system.

## What It Does

This tool helps you maintain a personal knowledge base by:

- Retrieving your highlights from Readdeck
- Organizing them by their parent article or document
- Generating structured Markdown notes in your Zettelkasten system
- Grouping highlights by color (with customizable meanings)
- Preserving metadata like source URLs, publication dates, and authors

## Key Features

- **Idempotent Operation**: Safely run the exporter multiple times without creating duplicates. Only new highlights are added to existing notes.
- **Read-Only Source**: Never modifies your Readdeck data; only reads from it.
- **Color Organization**: Highlights are grouped by color categories (customizable).
- **Metadata Preservation**: Keeps important context like source URL, publication date, and authors.
- **Content Preservation**: Keeps all changes to a document when it was originally exported with the tool
- **Configurable**: Simple configuration through CLI commands or configuration file.

## How To Use It
### Configuration Options

Set the required configuration: 
```
highlight-exporter config --base-url=$BASE_URL --token=$AUTH_TOKEN --fleeting-path=/home/user/notes/zettelkasten/fleeting
```

View your current configuration:
```
highlight-exporter config view
```

Update specific settings:
```
highlight-exporter config --timeout=45s --bookmarks-per-page=90
```

## TODO
- [x] Make exporter CLI command
- [ ] Save state of
    - [ ] Most recent highlight
    - [ ] Lookup for files with readdeck id, to skip reading the files. Limiting IO.
- [ ] Better logging
    - [ ] Log in the same line, while doing stuff (eg walking path, fetching api's, writing files)
    - [ ] Show which files needed updates, created, or NO-OP
    - [ ] Better summary, eg X Amount created, X Updated, X No 
    - [ ] Detailed (verbose) that shows the info that's now with the len 20 but without an if statement
    - [ ] A 'timed' flag, that times/benchmarks the run, tells you how long it was busy, maybe default to true, or without a flag?
- [ ] Make it available in my nix packages somehow?

## License
MIT
