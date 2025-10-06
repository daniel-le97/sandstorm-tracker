import { appendFile, mkdir } from "fs/promises";
import path from "path";

export type LogLevel = 'error' | 'warn' | 'info' | 'debug';

// Capture the native console methods once so logger internals stay stable even if
// global `console` is later hijacked to forward to this logger.
const nativeConsole = {
    log: console.log.bind( console ),
    info: console.info.bind( console ),
    warn: console.warn.bind( console ),
    error: console.error.bind( console ),
    debug: ( console.debug || console.log ).bind( console ),
};

let currentLevel: LogLevel = 'info';

const LEVELS: Record<LogLevel, number> = { error: 0, warn: 1, info: 2, debug: 3 };

export function setLogLevel ( level: LogLevel ) {
    currentLevel = level;
}

export function setVerbose ( verbose: boolean ) {
    setLogLevel( verbose ? 'debug' : 'info' );
}

function shouldLog ( level: LogLevel ) {
    return LEVELS[ level ] <= LEVELS[ currentLevel ];
}

// File logging state
let fileLoggingEnabled = true; // default enabled per user's request
let logFilePath = path.join( process.cwd(), "sandstorm-tracker.log" );

async function writeToFile ( msg: string ) {
    try
    {
        await mkdir( path.dirname( logFilePath ), { recursive: true } );
        // appendFile returns a Promise; we deliberately don't await here to avoid
        // slowing main thread — errors are caught and logged by nativeConsole
        appendFile( logFilePath, msg ).catch( ( e ) => nativeConsole.error( "Failed to write log file:", e ) );
    } catch ( e )
    {
        nativeConsole.error( "Failed to prepare log file:", e );
    }
}

export async function enableFileLogging ( enabled: boolean = true, filePath?: string ) {
    fileLoggingEnabled = Boolean( enabled );
    if ( filePath )
    {
        logFilePath = filePath;
    }
    if ( fileLoggingEnabled )
    {
        await mkdir( path.dirname( logFilePath ), { recursive: true } ).catch( () => { } );
        await appendFile( logFilePath, `--- Sandstorm Tracker log started: ${ new Date().toISOString() } ---\n` ).catch( ( e ) => nativeConsole.error( "Failed to open log file:", e ) );
    }
}

function formatArgs ( args: any[] ) {
    return args
        .map( ( a ) => {
            if ( typeof a === 'string' ) return a;
            try
            {
                return JSON.stringify( a );
            } catch
            {
                return String( a );
            }
        } )
        .join( ' ' );
}

export function error ( ...args: any[] ) {
    if ( shouldLog( 'error' ) )
    {
        nativeConsole.error( ...args );
        if ( fileLoggingEnabled )
        {
            const msg = `${ new Date().toISOString() } [ERROR] ${ formatArgs( args ) }\n`;
            void writeToFile( msg );
        }
    }
}

export function warn ( ...args: any[] ) {
    if ( shouldLog( 'warn' ) )
    {
        nativeConsole.warn( ...args );
        if ( fileLoggingEnabled )
        {
            const msg = `${ new Date().toISOString() } [WARN ] ${ formatArgs( args ) }\n`;
            void writeToFile( msg );
        }
    }
}

export function info ( ...args: any[] ) {
    if ( shouldLog( 'info' ) )
    {
        nativeConsole.info( ...args );
        if ( fileLoggingEnabled )
        {
            const msg = `${ new Date().toISOString() } [INFO ] ${ formatArgs( args ) }\n`;
            void writeToFile( msg );
        }
    }
}

export function debug ( ...args: any[] ) {
    if ( shouldLog( 'debug' ) )
    {
        nativeConsole.debug( ...args );
        if ( fileLoggingEnabled )
        {
            const msg = `${ new Date().toISOString() } [DEBUG] ${ formatArgs( args ) }\n`;
            void writeToFile( msg );
        }
    }
}

export default { setLogLevel, setVerbose, enableFileLogging, error, warn, info, debug };
