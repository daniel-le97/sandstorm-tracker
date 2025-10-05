import { initializeApplication, startWatching, stopApp } from "./src/app";

// Thin launcher: import and call the main application
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
        await initializeApplication();
        // install graceful shutdown handlers


        await startWatching();
    } catch ( error )
    {
        console.error( "❌ Failed to start multi-server tracker:", error );
        process.exit( 1 );
    }
}

main().catch( console.error );
