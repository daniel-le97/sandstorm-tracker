import { useRealtimeRecords } from '../../hooks/useRealtimeRecords';
import { usePB } from '../../hooks/usePB';
import { useState } from 'preact/hooks';
import styles from './MatchHistory.module.css';

const MAP_IMAGES = {
    'Crossing': 'crossing.png',
    'Farmhouse': 'farmhouse.png',
    'Hideout': 'hideout.png',
    'Precinct': 'precinct.png',
    'Refinery': 'refinery.png',
    'Summit': 'summit.png',
    'Outskirts': 'outskirts.jpg',
    'Ministry': 'ministry.png',
    'Hillside': 'hillside.png',
    'Power Plant': 'power_plant.jpg',
};

function getMapImage ( title ) {
    return MAP_IMAGES[ title ] || '';
}

interface MatchCardProps {
    match: any;
    mapImage: string;
}

function MatchCard ( { match, mapImage }: MatchCardProps ) {
    const pb = usePB();
    const isActive = !match.end_time;
    const [ isExpanded, setIsExpanded ] = useState( false );
    const [ players, setPlayers ] = useState<any[]>( [] );
    const [ loading, setLoading ] = useState( false );

    const handleExpand = async () => {
        if ( isExpanded )
        {
            setIsExpanded( false );
            return;
        }

        setLoading( true );
        try
        {
            const result = await pb.collection( 'match_player_stats' ).getList( 1, 100, {
                filter: `match = "${ match.id }"`,
                sort: '-score',
                expand: 'player',
            } );
            setPlayers( result.items || [] );
        } catch ( err )
        {
            console.error( 'Failed to load player stats', err );
        } finally
        {
            setLoading( false );
            setIsExpanded( true );
        }
    };

    return (
        <div key={ match.id } class={ styles.matchCard }>
            <button
                class={ styles.matchHeaderButton }
                onClick={ handleExpand }
                type="button"
            >
                <div class={ styles.matchHeader }>
                    { mapImage ? (
                        <img
                            src={ `/maps/${ mapImage }` }
                            alt={ match.title }
                            class={ styles.mapImage }
                        />
                    ) : (
                        <div class={ styles.noImage }>No Image</div>
                    ) }
                    <div class={ styles.matchBasicInfo }>
                        <h3>{ match.title }</h3>
                        <p><strong>Mode:</strong> { match.mode }</p>
                        <p><strong>Lighting:</strong> { match.lighting || 'Day' }</p>
                    </div>
                    <div class={ styles.matchTime }>
                        <p>
                            <strong>Started:</strong> { new Date( match.start_time ).toLocaleString() }
                        </p>
                        { match.end_time && (
                            <p>
                                <strong>Ended:</strong> { new Date( match.end_time ).toLocaleString() }
                            </p>
                        ) }
                        <span class={ isActive ? styles.active : styles.completed }>
                            { isActive ? 'ACTIVE' : 'COMPLETED' }
                        </span>
                    </div>
                    <div class={ styles.expandIcon }>
                        { isExpanded ? '▼' : '▶' }
                    </div>
                </div>
            </button>

            { isExpanded && (
                <div class={ styles.playerStats }>
                    { loading ? (
                        <p style="padding: 1rem; text-align: center;">Loading players...</p>
                    ) : players.length === 0 ? (
                        <p style="padding: 1rem; text-align: center;">No player data</p>
                    ) : (
                        <table>
                            <thead>
                                <tr>
                                    <th>Player</th>
                                    <th style="text-align: center">Score</th>
                                    <th style="text-align: center">K</th>
                                    <th style="text-align: center">D</th>
                                    <th style="text-align: center">A</th>
                                    <th style="text-align: center">K/D</th>
                                </tr>
                            </thead>
                            <tbody>
                                { players.map( ( stat ) => {
                                    const playerName = stat.expand?.player?.name || 'Unknown';
                                    const kills = stat.kills || 0;
                                    const deaths = stat.deaths || 0;
                                    const kd = deaths > 0 ? ( kills / deaths ).toFixed( 2 ) : kills > 0 ? kills : 0;

                                    return (
                                        <tr key={ stat.id }>
                                            <td>{ playerName }</td>
                                            <td style="text-align: center; color: #ffd700; font-weight: bold;">
                                                { stat.score || 0 }
                                            </td>
                                            <td style="text-align: center; color: #4caf50;">
                                                { kills }
                                            </td>
                                            <td style="text-align: center; color: #f44336;">
                                                { deaths }
                                            </td>
                                            <td style="text-align: center; color: #bbb;">
                                                { stat.assists || 0 }
                                            </td>
                                            <td style="text-align: center; color: #9c27b0;">
                                                { kd }
                                            </td>
                                        </tr>
                                    );
                                } ) }
                            </tbody>
                        </table>
                    ) }
                </div>
            ) }
        </div>
    );
}

export function MatchHistory () {
    const [ selectedServer, setSelectedServer ] = useState( '' );
    const [ selectedMap, setSelectedMap ] = useState( '' );
    const [ selectedMode, setSelectedMode ] = useState( '' );

    const { records: servers, loading: serversLoading } = useRealtimeRecords( 'servers' );
    const { records: matches } = useRealtimeRecords( 'matches', '', '-start_time' );

    // Get unique maps and modes
    const uniqueMaps = [ ...new Set( matches.map( ( m ) => m.title ) ) ].sort();
    const uniqueModes = [ ...new Set( matches.map( ( m ) => m.mode ) ) ].sort();

    // Filter matches
    let filtered = matches;
    if ( selectedServer )
    {
        filtered = filtered.filter( ( m ) => m.server === selectedServer );
    }
    if ( selectedMap )
    {
        filtered = filtered.filter( ( m ) => m.title === selectedMap );
    }
    if ( selectedMode )
    {
        filtered = filtered.filter( ( m ) => m.mode === selectedMode );
    }

    if ( serversLoading ) return <div>Loading...</div>;

    return (
        <div class={ styles.container }>
            <h1>Match History</h1>

            <div class={ styles.layout }>
                {/* Filters */ }
                <aside class={ styles.filters }>
                    <h3>Filters</h3>
                    <form>
                        <div class={ styles.filterGroup }>
                            <label>Server</label>
                            <select
                                value={ selectedServer }
                                onChange={ ( e ) => setSelectedServer( e.currentTarget.value ) }
                            >
                                <option value="">All Servers</option>
                                { servers.map( ( s ) => (
                                    <option key={ s.id } value={ s.id }>
                                        { s.name }
                                    </option>
                                ) ) }
                            </select>
                        </div>

                        <div class={ styles.filterGroup }>
                            <label>Map</label>
                            <select
                                value={ selectedMap }
                                onChange={ ( e ) => setSelectedMap( e.currentTarget.value ) }
                            >
                                <option value="">All Maps</option>
                                { uniqueMaps.map( ( m ) => (
                                    <option key={ m } value={ m }>
                                        { m }
                                    </option>
                                ) ) }
                            </select>
                        </div>

                        <div class={ styles.filterGroup }>
                            <label>Mode</label>
                            <select
                                value={ selectedMode }
                                onChange={ ( e ) => setSelectedMode( e.currentTarget.value ) }
                            >
                                <option value="">All Modes</option>
                                { uniqueModes.map( ( m ) => (
                                    <option key={ m } value={ m }>
                                        { m }
                                    </option>
                                ) ) }
                            </select>
                        </div>

                        <button type="reset" onClick={ () => {
                            setSelectedServer( '' );
                            setSelectedMap( '' );
                            setSelectedMode( '' );
                        } }>
                            Clear Filters
                        </button>
                    </form>
                </aside>

                {/* Match List */ }
                <div class={ styles.matchList }>
                    { filtered.length === 0 ? (
                        <p>No matches found</p>
                    ) : (
                        filtered.map( ( match ) => {
                            const mapImage = getMapImage( match.title );
                            return <MatchCard key={ match.id } match={ match } mapImage={ mapImage } />;
                        } )
                    ) }
                </div>
            </div>
        </div>
    );
}
