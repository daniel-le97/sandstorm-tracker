import { useState, useEffect } from 'preact/hooks';
import { usePB } from './usePB';

export function useRealtimeRecords ( collection, filter = '', sort = '-created' ) {
    const pb = usePB();
    const [ records, setRecords ] = useState( [] );
    const [ loading, setLoading ] = useState( true );
    const [ error, setError ] = useState( null );

    useEffect( () => {
        let unsubscribe;

        ( async () => {
            try
            {
                // Fetch initial data
                const data = await pb.collection( collection ).getList( 1, 50, {
                    filter,
                    sort,
                } );
                setRecords( data.items );
                setLoading( false );

                // Subscribe to real-time updates
                unsubscribe = await pb.collection( collection ).subscribe( '*', ( e ) => {
                    setRecords( ( prev ) => {
                        // Handle record updates
                        if ( e.action === 'create' )
                        {
                            return [ e.record, ...prev ];
                        } else if ( e.action === 'update' )
                        {
                            return prev.map( ( r ) => ( r.id === e.record.id ? e.record : r ) );
                        } else if ( e.action === 'delete' )
                        {
                            return prev.filter( ( r ) => r.id !== e.record.id );
                        }
                        return prev;
                    } );
                } );
            } catch ( err )
            {
                setError( err.message );
                setLoading( false );
            }
        } )();

        return () => {
            if ( unsubscribe ) unsubscribe();
        };
    }, [ collection, filter, sort ] );

    return { records, loading, error };
}
