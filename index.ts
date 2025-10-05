import { root } from './src/cli';
import { stopApp } from './src/app';

// Thin launcher: delegate to CLI (root command) while retaining graceful shutdown
async function main (): Promise<void> {
    process.on( 'SIGINT', async () => {
        console.log( 'Received SIGINT' );
        await stopApp();
        process.exit( 0 );
    } );
    process.on( 'SIGTERM', async () => {
        console.log( 'Received SIGTERM' );
        await stopApp();
        process.exit( 0 );
    } );

    try
    {
        // Delegate startup to the CLI root command (it will call initialize/startWatching when appropriate)
        await root.run();
    } catch ( error )
    {
        console.error( "❌ Failed to start multi-server tracker:", error );
        process.exit( 1 );
    }
}

main().catch( console.error );
