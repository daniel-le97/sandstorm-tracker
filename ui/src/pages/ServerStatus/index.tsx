import { useRealtimeRecords } from '../../hooks/useRealtimeRecords';
import { usePB } from '../../hooks/usePB';
import { useState, useEffect } from 'preact/hooks';
import styles from './ServerStatus.module.css';

interface ServerWithMatch {
    server: any;
    match?: any;
    players: any[];
}

export function ServerStatus () {
    const pb = usePB();
    const { records: servers, loading: serversLoading } = useRealtimeRecords( 'servers' );
    const { records: matches } = useRealtimeRecords( 'matches', 'end_time = ""', '-start_time' );
    const [ serverDetails, setServerDetails ] = useState<Record<string, ServerWithMatch>>( {} );

    useEffect( () => {
        const loadServerDetails = async () => {
            const details: Record<string, ServerWithMatch> = {};

            // Initialize all servers first
            servers.forEach( ( server ) => {
                details[ server.id ] = {
                    server,
                    players: [],
                };
            } );

            // Load player stats for each match
            for ( const match of matches )
            {
                try
                {
                    const players = await pb.collection( 'match_player_stats' ).getList( 1, 100, {
                        filter: `match = "${ match.id }"`,
                        sort: '-score',
                        expand: 'player',
                    } );

                    if ( match.server && details[ match.server ] )
                    {
                        details[ match.server ] = {
                            ...details[ match.server ],
                            match,
                            players: players.items || [],
                        };
                    }
                } catch ( err )
                {
                    console.error( 'Failed to load player stats for match', match.id, err );
                }
            }

            setServerDetails( details );
        };

        if ( servers.length > 0 )
        {
            loadServerDetails();
        }
    }, [ matches, servers ] );

    if ( serversLoading || servers.length === 0 ) return <div>Loading...</div>;

    return (
        <div class={ styles.container }>
            <h1>Server Status</h1>
            <div class={ styles.serverGrid }>
                { Object.values( serverDetails ).map( ( detail ) => {
                    const { server, match, players } = detail;
                    const isActive = !!match;

                    return (
                        <div key={ server.id } class={ styles.serverCard }>
                            <div class={ styles.cardHeader }>
                                <h2>{ server.name }</h2>
                                <span class={ `${ styles.status } ${ isActive ? styles.active : '' }` }>
                                    { isActive ? '● ACTIVE' : '○ NO MATCH' }
                                </span>
                            </div>

                            { isActive && match && (
                                <>
                                    <div class={ styles.matchInfo }>
                                        <p><strong>Map:</strong> { match.title }</p>
                                        <p><strong>Mode:</strong> { match.mode }</p>
                                        <p><strong>Players:</strong> { players.length }</p>
                                    </div>

                                    { players.length > 0 && (
                                        <div class={ styles.playersTable }>
                                            <table>
                                                <thead>
                                                    <tr>
                                                        <th>Player</th>
                                                        <th style="text-align: center">Score</th>
                                                        <th style="text-align: center">K</th>
                                                        <th style="text-align: center">D</th>
                                                        <th style="text-align: center">A</th>
                                                    </tr>
                                                </thead>
                                                <tbody>
                                                    { players.slice( 0, 10 ).map( ( stat ) => (
                                                        <tr key={ stat.id }>
                                                            <td>{ stat.expand?.player?.name || 'Unknown' }</td>
                                                            <td style="text-align: center; color: #ffd700;">
                                                                { stat.score || 0 }
                                                            </td>
                                                            <td style="text-align: center; color: #4caf50;">
                                                                { stat.kills || 0 }
                                                            </td>
                                                            <td style="text-align: center; color: #f44336;">
                                                                { stat.deaths || 0 }
                                                            </td>
                                                            <td style="text-align: center; color: #bbb;">
                                                                { stat.assists || 0 }
                                                            </td>
                                                        </tr>
                                                    ) ) }
                                                </tbody>
                                            </table>
                                        </div>
                                    ) }
                                </>
                            ) }
                        </div>
                    );
                } ) }
            </div>
        </div>
    );
}
