import { describe, it, expect } from 'bun:test';
import { Command } from '../src/cli/command';

async function captureAsync ( fn: () => Promise<void> ) {
    const logs: string[] = [];
    const old = console.log;
    // capture console.log calls as single-line strings
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    ( console as any ).log = ( ...args: any[] ) => logs.push( args.map( String ).join( ' ' ) );
    try
    {
        await fn();
    } finally
    {
        console.log = old;
    }
    return logs.join( '\n' );
}

function buildExampleCLI () {
    const root = new Command( { name: 'sandstorm', description: 'Sandstorm tracker CLI' } );

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
            if ( args.length ) console.log( 'extra args:', args.join( ' ' ) );
        },
    } );

    root.addCommand( serve );

    const version = new Command( { name: 'version', description: 'Show version', action: () => console.log( 'sandstorm 2.0.0' ) } );
    root.addCommand( version );

    root.action = ( { flags } ) => {
        if ( flags.verbose ) console.log( 'root verbose' );
        console.log( 'Run with `serve` or `version`. Use --help for details' );
    };

    return { root, serve, version };
}



describe( 'CLI integration', () => {
    it( 'parses boolean and string flags', () => {
        const cmd = new Command( { name: 'root', flags: { verbose: { type: 'boolean', alias: 'v', default: false }, port: { type: 'string', alias: 'p', default: '3000' } } } );
        const parsed = cmd.parse( [ '--verbose', '--port', '8080', 'positional1' ] );
        expect( parsed.flags.verbose ).toBe( true );
        expect( parsed.flags.port ).toBe( '8080' );
        expect( parsed.args ).toEqual( [ 'positional1' ] );
    } );
    it( 'prints root usage with --help', async () => {
        const { root } = buildExampleCLI();
        const out = await captureAsync( async () => await root.run( [ '--help' ] ) );
        expect( out ).toContain( 'Usage: sandstorm' );
        expect( out ).toContain( 'Commands:' );
        expect( out ).toContain( 'serve' );
        expect( out ).toContain( 'version' );
    } );

    it( 'help subcommand prints root usage and target command usage', async () => {
        const { root, serve } = buildExampleCLI();

        const rootHelp = await captureAsync( async () => await root.run( [ 'help' ] ) );
        expect( rootHelp ).toContain( 'Usage: sandstorm' );

        const serveHelp = await captureAsync( async () => await root.run( [ 'help', 'serve' ] ) );
        // serve usage should include the port flag and its default
        expect( serveHelp ).toContain( 'Usage: serve' );
        expect( serveHelp ).toContain( '--port' );
        expect( serveHelp ).toContain( '(default: "3000")' );
    } );

    it( 'serve runs and prints starting message with --port and alias -p', async () => {
        const { root } = buildExampleCLI();

        const out1 = await captureAsync( async () => await root.run( [ 'serve', '--port', '8080' ] ) );
        expect( out1 ).toContain( 'Starting server on port 8080' );

        const out2 = await captureAsync( async () => await root.run( [ 'serve', '-p', '4000' ] ) );
        expect( out2 ).toContain( 'Starting server on port 4000' );
    } );

    it( 'serve prints extra positional args', async () => {
        const { root } = buildExampleCLI();
        const out = await captureAsync( async () => await root.run( [ 'serve', '--port', '9000', 'alpha', 'beta' ] ) );
        expect( out ).toContain( 'Starting server on port 9000' );
        expect( out ).toContain( 'extra args: alpha beta' );
    } );

    it( 'version prints the version string', async () => {
        const { root } = buildExampleCLI();
        const out = await captureAsync( async () => await root.run( [ 'version' ] ) );
        expect( out ).toContain( 'sandstorm 2.0.0' );
    } );

    it( 'runs root default action when no subcommand provided', async () => {
        const { root } = buildExampleCLI();
        const out = await captureAsync( async () => await root.run( [] ) );
        expect( out ).toContain( 'Run with `serve` or `version`. Use --help for details' );
    } );
} );
