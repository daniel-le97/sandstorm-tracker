import { test, expect, describe, afterAll } from 'bun:test';
import { writeFileSync, unlinkSync, statSync } from 'fs';

describe( 'File Watcher', () => {
    const testFile = 'test-watcher-file.log';

    afterAll( () => {
        // Clean up test file
        try
        {
            unlinkSync( testFile );
        } catch ( e )
        {
            // Ignore cleanup errors
        }
    } );

    test( 'File creation and modification detection', () => {
        // Create initial file
        const initialContent = 'Initial log content\n';
        writeFileSync( testFile, initialContent );

        // Check file exists and has content
        expect( () => statSync( testFile ) ).not.toThrow();

        const content = Bun.file( testFile ).text();
        expect( content ).resolves.toBe( initialContent );
    } );

    test( 'File size change detection', () => {
        // Create file with known content
        const content1 = 'Line 1\n';
        writeFileSync( testFile, content1 );
        const size1 = statSync( testFile ).size;

        // Append content
        const content2 = content1 + 'Line 2\n';
        writeFileSync( testFile, content2 );
        const size2 = statSync( testFile ).size;

        // Verify size changed
        expect( size2 ).toBeGreaterThan( size1 );
        expect( size2 - size1 ).toBe( 'Line 2\n'.length );
    } );

    test( 'File content reading works correctly', async () => {
        const testContent = `
[2025.10.05-12.00.00:000][001]LogNet: Join succeeded: TestPlayer
[2025.10.05-12.01.00:000][002]LogGameplayEvents: Display: Game over
[2025.10.05-12.02.00:000][003]LogChat: Display: TestPlayer(12345) Global Chat: !stats
        `.trim();

        writeFileSync( testFile, testContent );

        // Read file content
        const fileContent = await Bun.file( testFile ).text();
        expect( fileContent ).toBe( testContent );

        // Check lines can be split correctly
        const lines = fileContent.split( '\n' ).filter( line => line.trim() );
        expect( lines ).toHaveLength( 3 );
        expect( lines[ 0 ] ).toContain( 'Join succeeded' );
        expect( lines[ 1 ] ).toContain( 'Game over' );
        expect( lines[ 2 ] ).toContain( '!stats' );
    } );

    test( 'Last line extraction works', async () => {
        const logLines = [
            '[2025.10.05-12.00.00:000][001]LogNet: Join succeeded: Player1',
            '[2025.10.05-12.01.00:000][002]LogNet: Join succeeded: Player2',
            '[2025.10.05-12.02.00:000][003]LogGameplayEvents: Display: Game over'
        ];

        writeFileSync( testFile, logLines.join( '\n' ) );

        const content = await Bun.file( testFile ).text();
        const lines = content.split( '\n' );
        const lastLine = lines.filter( line => line.trim() ).pop();

        expect( lastLine ).toBe( logLines[ 2 ] );
        expect( lastLine ).toContain( 'Game over' );
    } );

    test( 'Empty file handling', async () => {
        writeFileSync( testFile, '' );

        const content = await Bun.file( testFile ).text();
        expect( content ).toBe( '' );

        const lines = content.split( '\n' );
        const lastLine = lines.filter( line => line.trim() ).pop();
        expect( lastLine ).toBeUndefined();
    } );

    test( 'File with only whitespace handling', async () => {
        writeFileSync( testFile, '   \n\t\n   \n' );

        const content = await Bun.file( testFile ).text();
        const lines = content.split( '\n' );
        const lastLine = lines.filter( line => line.trim() ).pop();

        expect( lastLine ).toBeUndefined();
    } );

    test( 'Large file handling', async () => {
        // Create a file with many lines
        const lines = [];
        for ( let i = 0; i < 1000; i++ )
        {
            lines.push( `[2025.10.05-12.${ i.toString().padStart( 2, '0' ) }.00:000][${ i.toString().padStart( 3, '0' ) }]LogNet: Test line ${ i }` );
        }

        writeFileSync( testFile, lines.join( '\n' ) );

        const content = await Bun.file( testFile ).text();
        const contentLines = content.split( '\n' );
        const lastLine = contentLines.filter( line => line.trim() ).pop();

        expect( lastLine ).toBe( lines[ 999 ] );
        expect( contentLines.length ).toBe( 1000 );
    } );

    test( 'File append simulation', async () => {
        // Start with initial content
        const initialLines = [
            '[2025.10.05-12.00.00:000][001]LogNet: Join succeeded: Player1',
            '[2025.10.05-12.01.00:000][002]LogNet: Join succeeded: Player2'
        ];

        writeFileSync( testFile, initialLines.join( '\n' ) );

        // Simulate appending new lines (like a real log file)
        const newLines = [
            '[2025.10.05-12.02.00:000][003]LogGameplayEvents: Display: Game over',
            '[2025.10.05-12.03.00:000][004]LogChat: Display: Player1(12345) Global Chat: !stats'
        ];

        const allLines = [ ...initialLines, ...newLines ];
        writeFileSync( testFile, allLines.join( '\n' ) );

        const content = await Bun.file( testFile ).text();
        const lines = content.split( '\n' ).filter( line => line.trim() );

        expect( lines ).toHaveLength( 4 );
        expect( lines[ 2 ] ).toContain( 'Game over' );
        expect( lines[ 3 ] ).toContain( '!stats' );

        // Test getting just the last line
        const lastLine = lines.pop();
        expect( lastLine ).toContain( '!stats' );
    } );
} );