import type { Action } from '../command';
import { getStatements } from '../../database';

export const logFilesAction: Action = async ( { flags } ) => {
    try
    {
        const where: string[] = [];
        const params: any[] = [];
        if ( flags.server )
        {
            where.push( 'server_id = (SELECT id FROM servers WHERE server_id = ?)' );
            params.push( flags.server );
        }
        if ( flags.path )
        {
            where.push( 'log_path LIKE ?' );
            params.push( `%${ flags.path }%` );
        }
        const whereClause = where.length ? `WHERE ${ where.join( ' AND ' ) }` : '';
        const limit = parseInt( flags.limit || '20', 10 ) || 20;
        const sql = `SELECT lf.id, s.server_id as server_uuid, lf.log_path, lf.open_time, lf.lines_processed, lf.file_size_bytes, lf.created_at, lf.updated_at
                     FROM log_files lf
                     JOIN servers s ON lf.server_id = s.id
                     ${ whereClause }
                     ORDER BY lf.open_time DESC
                     LIMIT ${ limit }`;
        const db = ( await import( '../../database' ) ).DB.getDefaultDatabase();
        const rows = db.query( sql ).all( ...params ) as any[];
        if ( !rows.length )
        {
            console.log( 'No log file entries found.' );
            return;
        }
        console.table( rows );
    } catch ( e )
    {
        console.error( 'Failed to list log files:', e );
        process.exitCode = 3;
    }
};
