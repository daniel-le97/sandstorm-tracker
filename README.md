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

This project was created using `bun init` in bun v1.2.22. [Bun](https://bun.com) is a fast all-in-one JavaScript runtime.
