#!/usr/bin/env bun
import { Command } from './lib/index';
// Index only wires up command definitions and imports their actions.
import { setVerbose } from '../lib/console/logger';
import { serveAction } from './commands/serve';
import { versionAction } from './commands/version';
import { updateAction } from './commands/update';
import { logFilesAction } from './commands/log-files';

export const root = new Command( {
    name: 'sandstorm',
    description: 'Sandstorm tracker CLI',
} );

root.flags = {
    verbose: { type: 'boolean', alias: 'v', description: 'Enable verbose logging', default: false },
};

// set verbosity based on env/flags at runtime in index.ts before run()

const serve = new Command( {
    name: 'serve',
    description: 'Start server',
    flags: {
        port: { type: 'string', alias: 'p', description: 'Port to listen on', default: '3000' },
    },
    action: serveAction,
} );

root.addCommand( serve );

const version = new Command( { name: 'version', description: 'Show version', action: versionAction } );
root.addCommand( version );

// update subcommand: check, download, extract, optional install instructions
const update = new Command( {
    name: 'update',
    description: 'Check for updates and optionally download/extract the latest release',
    flags: {
        check: { type: 'boolean', description: 'Only check for updates (default)', default: true },
        download: { type: 'boolean', description: 'Download and extract the latest release', default: false },
        install: { type: 'boolean', description: 'After download, install the binary (unsafe if not used carefully)', default: false },
        yes: { type: 'boolean', description: 'Auto-confirm prompts (use with care)', default: false },
        target: { type: 'string', description: 'Install target path (overrides default)', default: undefined },
        outDir: { type: 'string', description: 'Directory to download/extract into', default: undefined },
        repo: { type: 'string', description: 'GitHub repo owner/name (owner/repo)', default: undefined },
    },
    action: updateAction,
} );

root.addCommand( update );

// log-files subcommand: list tracked log file entries per server or all
const logFiles = new Command( {
    name: 'log-files',
    description: 'List tracked log file metadata',
    flags: {
        server: { type: 'string', description: 'Filter by server UUID (server_id column)', default: undefined },
        limit: { type: 'string', description: 'Limit number of rows (default 20)', default: '20' },
        path: { type: 'string', description: 'Filter by log path contains', default: undefined },
    },
    action: logFilesAction,
} );

root.addCommand( logFiles );

// default action if no subcommand
root.action = ( { flags } ) => {
    if ( flags.verbose ) console.log( 'root verbose' );
    console.log( 'Run with `serve` or `version`. Use --help for details' );
};
