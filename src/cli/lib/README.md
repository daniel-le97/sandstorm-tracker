Quick CLI library for Bun (inspired by Cobra)

Usage

- Create a root Command and register subcommands with `addCommand`.
- Define `flags` for commands as an object of the form `{ name: { type: 'string'|'boolean', alias?: 'x', description?: string, default?: any } }`.
- Call `await root.run()` in your CLI entrypoint (the library handles `process.argv` by default).

Example

See `examples/cli.ts` for a minimal example with `serve` and `version` commands.

Notes

- This is a minimal, dependency-free library intended for small CLIs. It implements basic flag parsing, subcommands, and help output. It does not aim to cover the entire feature set of cobra.
