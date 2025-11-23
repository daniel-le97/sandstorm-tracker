import { useRealtimeRecords } from '../../hooks/useRealtimeRecords';
import { usePB } from '../../hooks/usePB';
import { useState, useEffect } from 'preact/hooks';
import styles from './Weapons.module.css';

interface WeaponStat {
    id: string;
    weapon_name: string;
    kills: number;
    assists: number;
}

function ServerWeaponsSection ( { server }: { server: any; } ) {
    const pb = usePB();
    const [ weapons, setWeapons ] = useState<WeaponStat[]>( [] );
    const [ loading, setLoading ] = useState( true );

    useEffect( () => {
        const loadWeapons = async () => {
            try
            {
                // Get all matches for this server
                const matches = await pb.collection( 'matches' ).getList( 1, 1000, {
                    filter: `server = "${ server.id }"`,
                } );

                if ( matches.items && matches.items.length > 0 )
                {
                    const matchIds = matches.items.map( ( m: any ) => m.id );
                    const matchFilters = matchIds.map( ( id: string ) => `match = "${ id }"` ).join( ' || ' );

                    // Get all weapon stats for matches on this server
                    const weaponStats = await pb.collection( 'match_weapon_stats' ).getList( 1, 1000, {
                        filter: matchFilters,
                    } );

                    // Aggregate by weapon name
                    const weaponMap: Record<string, WeaponStat> = {};
                    ( weaponStats.items || [] ).forEach( ( stat: any ) => {
                        const key = stat.weapon_name || 'Unknown';
                        if ( !weaponMap[ key ] )
                        {
                            weaponMap[ key ] = {
                                id: key,
                                weapon_name: key,
                                kills: 0,
                                assists: 0,
                            };
                        }
                        weaponMap[ key ].kills += stat.kills || 0;
                        weaponMap[ key ].assists += stat.assists || 0;
                    } );

                    const sorted = Object.values( weaponMap ).sort( ( a, b ) => b.kills - a.kills );
                    setWeapons( sorted );
                }
            } catch ( err )
            {
                console.error( 'Failed to load weapon stats', err );
            } finally
            {
                setLoading( false );
            }
        };

        loadWeapons();
    }, [ server ] );

    return (
        <div class={ styles.serverSection }>
            <h3>{ server.name }</h3>
            { loading ? (
                <p>Loading...</p>
            ) : weapons.length === 0 ? (
                <p>No weapon data</p>
            ) : (
                <table class={ styles.weaponsTable }>
                    <thead>
                        <tr>
                            <th>Weapon</th>
                            <th style="text-align: center">Kills</th>
                            <th style="text-align: center">Assists</th>
                        </tr>
                    </thead>
                    <tbody>
                        { weapons.map( ( weapon ) => (
                            <tr key={ weapon.id }>
                                <td>{ weapon.weapon_name }</td>
                                <td style="text-align: center; color: #4caf50;">{ weapon.kills }</td>
                                <td style="text-align: center; color: #bbb;">{ weapon.assists }</td>
                            </tr>
                        ) ) }
                    </tbody>
                </table>
            ) }
        </div>
    );
}

export function Weapons () {
    const { records: servers, loading: serversLoading } = useRealtimeRecords( 'servers' );

    if ( serversLoading ) return <div>Loading...</div>;

    return (
        <div class={ styles.container }>
            <h1>Weapon Usage</h1>

            <div class={ styles.serversList }>
                { servers.length === 0 ? (
                    <p>No servers found</p>
                ) : (
                    servers.map( ( server ) => (
                        <ServerWeaponsSection key={ server.id } server={ server } />
                    ) )
                ) }
            </div>
        </div>
    );
}
