#!/usr/bin/env bun
import { Command } from './cli/command';
import { initializeApplication, startWatching } from './app';

export const root = new Command( {
    name: 'sandstorm',
    description: 'Sandstorm tracker CLI',
} );

root.flags = {
    verbose: { type: 'boolean', alias: 'v', description: 'Enable verbose logging', default: false },
};

const serve = new Command( {
    name: 'serve',
    description: 'Start server',
    flags: {
        port: { type: 'string', alias: 'p', description: 'Port to listen on', default: '3000' },
    },
    action: async ( { flags, args } ) => {
        // start the actual application (initialize + start watchers)
        if ( flags.verbose ) console.log( 'verbose enabled' );
        if ( args.length ) console.log( 'extra args:', args );

        await initializeApplication();
        // startWatching will block (it awaits the watcher promises)
        await startWatching();
    },
} );

root.addCommand( serve );

const version = new Command( { name: 'version', description: 'Show version', action: () => console.log( 'sandstorm 2.0.0' ) } );
root.addCommand( version );

// default action if no subcommand
root.action = ( { flags } ) => {
    if ( flags.verbose ) console.log( 'root verbose' );
    console.log( 'Run with `serve` or `version`. Use --help for details' );
};
