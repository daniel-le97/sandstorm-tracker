#!/usr/bin/env bun
import { rmSync, existsSync } from 'fs';
import { resolve } from 'path';

const testDbDir = resolve( process.cwd(), 'tests/databases' );

console.log( '🧹 Cleaning up test databases...' );

if ( existsSync( testDbDir ) )
{
    try
    {
        rmSync( testDbDir, { recursive: true, force: true } );
        console.log( '✅ Test databases directory removed successfully' );
    } catch ( error )
    {
        console.error( '❌ Error removing test databases:', error );
        process.exit( 1 );
    }
} else
{
    console.log( '✅ Test databases directory doesn\'t exist - nothing to clean' );
}

console.log( '🎉 Cleanup complete!' );