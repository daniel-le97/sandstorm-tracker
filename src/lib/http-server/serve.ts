// Minimal Bun HTTP server wrapper
// Exports startServer and stopServer to be used by CLI or tests

type ServerHandle = {
    stop: () => Promise<void>;
};

let server: ServerHandle | null = null;

export interface ServeOptions {
    port?: number;
    hostname?: string;
}

export async function startServer ( opts: ServeOptions = {} ): Promise<void> {
    if ( server ) return; // already running

    const port = ( opts.port ?? Number( process.env.PORT ) ) || 8787;
    const hostname = opts.hostname ?? 'localhost';

    console.log( `Starting HTTP server on http://${ hostname }:${ port }` );

    const instance = Bun.serve( {
        hostname,
        port,
        async fetch ( req: Request ) {
            try
            {
                const url = new URL( req.url );
                if ( req.method === 'GET' && url.pathname === '/' )
                {
                    const body = {
                        status: 'ok',
                        pid: process.pid,
                        uptimeMs: Math.floor( process.uptime() * 1000 ),
                    };
                    return new Response( JSON.stringify( body ), { status: 200, headers: { 'Content-Type': 'application/json' } } );
                }

                if ( req.method === 'GET' && url.pathname === '/health' )
                {
                    return new Response( 'OK', { status: 200, headers: { 'Content-Type': 'text/plain' } } );
                }

                // Allow programmatic shutdown for tests/dev only
                if ( req.method === 'POST' && url.pathname === '/shutdown' )
                {
                    // stop after responding
                    setTimeout( () => {
                        stopServer().catch( ( e ) => console.error( 'Error stopping server:', e ) );
                    }, 10 );
                    return new Response( 'shutting down', { status: 200 } );
                }

                return new Response( 'Not Found', { status: 404 } );
            } catch ( err )
            {
                console.error( 'HTTP server handler error:', err );
                return new Response( 'Internal Server Error', { status: 500 } );
            }
        }
    } );

    // Wrap the returned object so tests can call stopServer
    server = {
        stop: async () => {
            try
            {
                // Bun's serve object has a stop method; call it if present
                if ( typeof ( instance as any ).stop === 'function' )
                {
                    await ( instance as any ).stop();
                }
            } finally
            {
                server = null;
            }
        }
    };

    // Log when server closes (if Bun provides this info)
    if ( typeof ( instance as any ).on === 'function' )
    {
        try
        {
            ( instance as any ).on( 'close', () => console.log( 'HTTP server closed' ) );
        } catch { /* ignore */ }
    }
}

export async function stopServer (): Promise<void> {
    if ( !server ) return;
    await server.stop();
}

export default { startServer, stopServer };
