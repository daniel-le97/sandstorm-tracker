#!/usr/bin/env bun
import { Command } from '../src/lib/cli/lib/command';

const root = new Command( {
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
        console.log( 'Starting server on port', flags.port );
        if ( flags.verbose ) console.log( 'verbose enabled' );
        if ( args.length ) console.log( 'extra args:', args );
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

await root.run();
