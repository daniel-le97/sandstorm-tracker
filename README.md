# sandstorm-wrapper

To install dependencies:

```bash
bun install
```

To run:

```bash
bun run index.ts
```

## Testing

To run tests:

```bash
bun test
```

To clean up test databases:

```bash
bun run test:cleanup
```

This removes all test database files in the `tests/databases` directory. Useful for cleaning up after test runs or when test databases become corrupted.

## Project Structure

```
├── index.ts              # Main application entry point
├── src/                  # Source code
│   ├── database.ts       # Database connection and schema
│   ├── events.ts         # Event parsing logic
│   ├── stats-service.ts  # Statistics service
│   ├── command-handler.ts # Chat command handler
│   ├── lib/              # Utility libraries
│   └── examples/         # Example code
├── tests/                # Test files
├── scripts/              # Build and utility scripts
├── docs/                 # Documentation files
├── config/               # Configuration files
└── logs/                 # Log files
```

This project was created using `bun init` in bun v1.2.22. [Bun](https://bun.com) is a fast all-in-one JavaScript runtime.
