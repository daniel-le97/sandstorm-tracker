import { root } from './src/cli';
import { stopApp } from './src/app';
import logger from './src/lib/console/logger';
import { hijackConsole } from './src/lib/console/console-hijack';
import { startServer, stopServer } from './src/lib/http-server/serve';
import appEmitter from './src/lib/emitter/emitter';

appEmitter.on("shutdown", () => {
    logger.info("App is shutting down, performing cleanup...");
});

// Thin launcher: delegate to CLI (root command) while retaining graceful shutdown
async function main (): Promise<void> {
    process.on( 'SIGINT', async () => {
        logger.info( 'Received SIGINT' );
        // appEmitter.emit("shutdown");
        await stopApp();
        await stopServer()
        process.exit( 0 );
    } );
    process.on( 'SIGTERM', async () => {
        logger.info( 'Received SIGTERM' );
        // appEmitter.emit("shutdown");
        await stopApp();
        await stopServer()
        process.exit( 0 );
    } );

    try
    {
        // Hijack global console methods to route through our logger early
        hijackConsole();

        // Respect env var for initial verbosity
        const envVerbose = process.env.SANDSTORM_VERBOSE === '1' || process.env.DEBUG === '1';
        logger.setVerbose?.( envVerbose );

        // Delegate startup to the CLI root command (it will call initialize/startWatching when appropriate)
        await root.run();
        await startServer()
    } catch ( error )
    {
        logger.error( "❌ Failed to start multi-server tracker:", error );
        process.exit( 1 );
    }
}

main().catch( console.error );
