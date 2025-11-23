import { useRealtimeRecords } from '../../hooks/useRealtimeRecords';
import { usePB } from '../../hooks/usePB';
import { useState, useEffect } from 'preact/hooks';
import styles from './Players.module.css';

interface PlayerStats {
    total_kills: number;
    total_deaths: number;
    total_assists: number;
    total_score: number;
    total_duration_seconds: number;
}

function PlayerCard ( { player, onSelect }: { player: any; onSelect: ( player: any ) => void; } ) {
    return (
        <button
            class={ styles.playerCard }
            onClick={ () => onSelect( player ) }
            type="button"
        >
            <h3>{ player.name }</h3>
            <p>
                <small class={ styles.steamId }>{ player.external_id }</small>
            </p>
        </button>
    );
}

function PlayerStatsModal ( { player, onClose }: { player: any; onClose: () => void; } ) {
    const pb = usePB();
    const [ stats, setStats ] = useState<PlayerStats | null>( null );
    const [ topWeapons, setTopWeapons ] = useState<any[]>( [] );
    const [ loading, setLoading ] = useState( true );

    useEffect( () => {
        const loadStats = async () => {
            try
            {
                const result = await pb.collection( 'player_total_stats' ).getOne( player.id );
                setStats( {
                    total_kills: result.total_kills || 0,
                    total_deaths: result.total_deaths || 0,
                    total_assists: result.total_assists || 0,
                    total_score: result.total_score || 0,
                    total_duration_seconds: result.total_duration_seconds || 0,
                } );

                // Load top 3 weapons
                const weapons = await pb.collection( 'player_weapon_stats' ).getList( 1, 100, {
                    filter: `player = "${ player.id }"`,
                    sort: '-total_kills',
                } );

                setTopWeapons( ( weapons.items || [] ).slice( 0, 3 ) );
            } catch ( err )
            {
                console.error( 'Failed to load player stats', err );
            } finally
            {
                setLoading( false );
            }
        };

        loadStats();
    }, [ player ] );

    const handleBackdropClick = ( e: MouseEvent ) => {
        if ( e.target === e.currentTarget )
        {
            onClose();
        }
    };

    const kd = stats && stats.total_deaths > 0 ? ( stats.total_kills / stats.total_deaths ).toFixed( 2 ) : stats?.total_kills || 0;
    const playTimeHours = stats ? ( stats.total_duration_seconds / 3600 ).toFixed( 1 ) : 0;
    const playTimeMinutes = stats ? Math.floor( stats.total_duration_seconds / 60 ) : 0;
    const scorePerMinute = stats && playTimeMinutes > 0 ? ( stats.total_score / playTimeMinutes ).toFixed( 2 ) : 0;

    return (
        <div class={ styles.modal } onClick={ handleBackdropClick }>
            <div class={ styles.modalContent }>
                <button class={ styles.closeBtn } onClick={ onClose } type="button">
                    ✕
                </button>
                <h2>{ player.name }</h2>
                <p class={ styles.steamId }>Steam ID: { player.external_id }</p>

                { loading ? (
                    <p>Loading stats...</p>
                ) : stats ? (
                    <>
                        <div class={ styles.statsGrid }>
                            <div class={ styles.statItem }>
                                <div class={ styles.statLabel }>Total Kills</div>
                                <div class={ styles.statValue } style="color: #4caf50;">{ stats.total_kills }</div>
                            </div>
                            <div class={ styles.statItem }>
                                <div class={ styles.statLabel }>Total Deaths</div>
                                <div class={ styles.statValue } style="color: #f44336;">{ stats.total_deaths }</div>
                            </div>
                            <div class={ styles.statItem }>
                                <div class={ styles.statLabel }>Total Assists</div>
                                <div class={ styles.statValue } style="color: #bbb;">{ stats.total_assists }</div>
                            </div>
                            <div class={ styles.statItem }>
                                <div class={ styles.statLabel }>K/D Ratio</div>
                                <div class={ styles.statValue } style="color: #9c27b0;">{ kd }</div>
                            </div>
                            <div class={ styles.statItem }>
                                <div class={ styles.statLabel }>Total Score</div>
                                <div class={ styles.statValue } style="color: #ffd700;">{ stats.total_score }</div>
                            </div>
                            <div class={ styles.statItem }>
                                <div class={ styles.statLabel }>Score/Min</div>
                                <div class={ styles.statValue } style="color: #ff9800;">{ scorePerMinute }</div>
                            </div>
                            <div class={ styles.statItem }>
                                <div class={ styles.statLabel }>Play Time</div>
                                <div class={ styles.statValue }>{ playTimeHours }h</div>
                            </div>
                        </div>

                        { topWeapons.length > 0 && (
                            <div class={ styles.topWeapons }>
                                <h4>Top 3 Weapons</h4>
                                <ul>
                                    { topWeapons.map( ( weapon, idx ) => (
                                        <li key={ idx }>
                                            <span class={ styles.weaponName }>{ weapon.weapon_name || 'Unknown' }</span>
                                            <span class={ styles.weaponKills }>{ weapon.total_kills || 0 } kills</span>
                                        </li>
                                    ) ) }
                                </ul>
                            </div>
                        ) }
                    </>
                ) : (
                    <p>No stats available</p>
                ) }
            </div>
        </div>
    );
}

export function Players () {
    const [ searchQuery, setSearchQuery ] = useState( '' );
    const [ sortBy, setSortBy ] = useState( 'name' );
    const [ selectedPlayer, setSelectedPlayer ] = useState<any>( null );

    const filter = searchQuery ? `name ~ '${ searchQuery }'` : '';
    const { records: players, loading } = useRealtimeRecords( 'players', filter, sortBy );

    if ( loading ) return <div>Loading...</div>;

    return (
        <div class={ styles.container }>
            <h1>Players</h1>

            <div class={ styles.controls }>
                <input
                    type="text"
                    placeholder="Search players..."
                    value={ searchQuery }
                    onChange={ ( e ) => setSearchQuery( e.currentTarget.value ) }
                    class={ styles.searchInput }
                />
                <select value={ sortBy } onChange={ ( e ) => setSortBy( e.currentTarget.value ) }>
                    <option value="name">Name (A-Z)</option>
                    <option value="-kills">Most Kills</option>
                    <option value="-deaths">Most Deaths</option>
                </select>
            </div>

            <div class={ styles.playerGrid }>
                { players.length === 0 ? (
                    <p>No players found</p>
                ) : (
                    players.map( ( player ) => (
                        <PlayerCard
                            key={ player.id }
                            player={ player }
                            onSelect={ setSelectedPlayer }
                        />
                    ) )
                ) }
            </div>

            { selectedPlayer && (
                <PlayerStatsModal
                    player={ selectedPlayer }
                    onClose={ () => setSelectedPlayer( null ) }
                />
            ) }
        </div>
    );
}
