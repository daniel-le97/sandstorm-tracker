// @ts-ignore
import Database from 'bun:sqlite';

const db = new Database( 'sandstorm-tracker.db' );

// const playerId = process.argv[ 2 ];
// if ( !playerId )
// {
//     console.error( 'Usage: bun script.ts <player_id>' );
//     process.exit( 1 );
// }

const rows = db.query(
    'SELECT id, killer_id, server_id, kill_type FROM kills WHERE killer_id = ? AND server_id = 1;'
).all( 1 );


for ( const row of rows )
{
    console.log( row );
}
console.log( `Total: ${ rows.length }` );
