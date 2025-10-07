import logger from "./logger";

// Provide a safe hijack of the global `console` methods so that existing
// console.log/warn/error/info/debug calls are routed through our logger.
// The logger module itself uses the original native console methods, so
// this wrapper avoids recursion.
export function hijackConsole (): void {
    // Avoid double-hijack
    if ( ( console as any ).__hijacked_by_sandstorm ) return;

    const wrapped: Partial<Console> = {
        log: ( ...args: any[] ) => logger.info( ...args ),
        info: ( ...args: any[] ) => logger.info( ...args ),
        warn: ( ...args: any[] ) => logger.warn( ...args ),
        error: ( ...args: any[] ) => logger.error( ...args ),
        debug: ( ...args: any[] ) => logger.debug( ...args ),
    };

    // Replace only the common methods; leave others untouched
    console.log = wrapped.log as any;
    console.info = wrapped.info as any;
    console.warn = wrapped.warn as any;
    console.error = wrapped.error as any;
    console.debug = wrapped.debug as any;

    // Mark the console as hijacked so we don't reapply
    ( console as any ).__hijacked_by_sandstorm = true;
}

export default { hijackConsole };
