# saga-courier-local-file

Appends tales to a local file in markdown format. Useful for generating a changelog file, feeding into a static site, or piping into other tools.

Each tale is written as a level-2 heading (the title) followed by the body text and a version footer. Tales are appended in delivery order; the file is created if it does not exist.

## Configuration

### Route config (in `.saga/config.jsonnet`)

No route config keys. The output file is set via environment variable so it stays out of version-controlled config.

### Environment variables

| Variable | Description |
|----------|-------------|
| `SAGA_COURIER_LOCAL_FILE__PATH` | Path to the file to append to. Created if it does not exist. Parent directory must exist. |
