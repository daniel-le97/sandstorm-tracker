import { describe, it, expect } from 'bun:test';
import { readFile, rm } from 'fs/promises';
import path from 'path';

describe( 'Logger integration', () => {
    it( 'writes to file and respects log level, and routes to native console', async () => {
        const tmp = path.join( process.cwd(), 'tests', 'tmp-logger.log' );
        // remove old file if present
        await rm( tmp, { force: true } ).catch( () => { /* ignore */ } );

        // Capture native console methods BEFORE importing the logger so that
        // the logger's internal nativeConsole binds to these wrappers.
        const captured: string[] = [];
        const oldConsole = {
            log: console.log,
            info: console.info,
            warn: console.warn,
            error: console.error,
            debug: console.debug,
        };

        // override console methods to capture single-line strings
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        ( console as any ).info = ( ...args: any[] ) => captured.push( [ 'INFO', ...args ].map( String ).join( ' ' ) );
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        ( console as any ).warn = ( ...args: any[] ) => captured.push( [ 'WARN', ...args ].map( String ).join( ' ' ) );
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        ( console as any ).error = ( ...args: any[] ) => captured.push( [ 'ERROR', ...args ].map( String ).join( ' ' ) );
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        ( console as any ).debug = ( ...args: any[] ) => captured.push( [ 'DEBUG', ...args ].map( String ).join( ' ' ) );

        try
        {
            // dynamically import so nativeConsole binds to our patched console
            const logger = await import( '../src/lib/logger' );

            // enable file logging to our temp path
            await logger.enableFileLogging( true, tmp );

            // set log level to 'warn' so info/debug do not get logged
            logger.setLogLevel( 'warn' );

            // emit various levels
            logger.info( 'this is info', 1 );
            logger.warn( 'this is warn', 2 );
            logger.error( 'this is error', 3 );
            logger.debug( 'this is debug', 4 );

            // allow async appendFile to complete (writes are fire-and-forget)
            await new Promise( ( r ) => setTimeout( r, 200 ) );

            const content = await readFile( tmp, 'utf8' );
            expect( content ).toContain( 'Sandstorm Tracker log started' );
            // warn and error should be present in file
            expect( content ).toContain( '[WARN ] this is warn 2' );
            expect( content ).toContain( '[ERROR] this is error 3' );
            // info and debug should not be present due to log level
            expect( content ).not.toContain( '[INFO ] this is info 1' );
            expect( content ).not.toContain( '[DEBUG] this is debug 4' );

            // also ensure our captured native console saw warn/error but not info/debug
            expect( captured.some( ( s ) => s.includes( 'this is warn' ) ) ).toBe( true );
            expect( captured.some( ( s ) => s.includes( 'this is error' ) ) ).toBe( true );
            expect( captured.some( ( s ) => s.includes( 'this is info' ) ) ).toBe( false );
            expect( captured.some( ( s ) => s.includes( 'this is debug' ) ) ).toBe( false );

            // cleanup file
            await rm(tmp, { force: true }).catch(() => { /* ignore */ });
        } finally
        {
            // restore original console methods
            console.log = oldConsole.log;
            console.info = oldConsole.info;
            console.warn = oldConsole.warn;
            console.error = oldConsole.error;
            console.debug = oldConsole.debug;
        }
    } );
} );
