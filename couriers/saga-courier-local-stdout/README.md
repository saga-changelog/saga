# saga-courier-local-stdout

Prints tales to stdout. Useful for local development, dry-run style testing, and verifying that a chapter's routing lines up the way you expect before wiring real couriers.

Saga prefixes each line of a courier's stdout with the courier name, so running `saga tell` with this courier gives you a readable transcript of what would go where.

## Configuration

### Route config (in `.saga/config.jsonnet`)

| Key | Description |
|-----|-------------|
| `prefix` | Optional label printed on each output line, handy when several routes share this courier and you want to tell them apart in the output. |

### Environment variables

| Variable | Description |
|----------|-------------|
| `SAGA_COURIER_LOCAL_STDOUT__NOTE` | Required. Any non-empty value. Exists so the courier can demonstrate saga's per-courier environment variable injection. |

The note value is printed in the output header so you can confirm the variable made it through saga's env-var filtering.
