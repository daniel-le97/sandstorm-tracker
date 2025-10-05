import { Command } from '../src/cli/command'
import { describe, it, expect } from 'bun:test'

describe('CLI parsing', () => {
	it('parses boolean and string flags', () => {
		const cmd = new Command({ name: 'root', flags: { verbose: { type: 'boolean', alias: 'v', default: false }, port: { type: 'string', alias: 'p', default: '3000' } } })
		const parsed = cmd.parse(['--verbose', '--port', '8080', 'positional1'])
		expect(parsed.flags.verbose).toBe(true)
		expect(parsed.flags.port).toBe('8080')
		expect(parsed.args).toEqual(['positional1'])
	})
})
