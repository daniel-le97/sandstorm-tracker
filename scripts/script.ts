//@ts-ignore
let content = ( await Bun.file( 'events/hc.log' ).text() ).split( '\n' );
let count = 0;
let rabbitKills = 0;
let bearKills = 0;
let blueKills = 0;
let originKills = 0;

let rabbitDeaths = 0;
let bearDeaths = 0;
let blueDeaths = 0;
let originDeaths = 0;

const players = [ 'Rabbit', 'Bear', 'Blue', '0rigin' ];
for ( const line of content )
{
    if ( !line.includes( 'killed' ) ) continue;

    const splitKill = line.split( 'killed' );
    // Split the killer section by ' + ' for multi-kills
    const killers = splitKill[ 0 ].split( ' + ' );
    for ( const killer of killers )
    {
        for ( const player of players )
        {
            if ( killer.includes( player ) )
            {
                count++;
                if ( player === 'Rabbit' )
                {
                    rabbitKills++;
                } else if ( player === 'Bear' )
                {
                    bearKills++;
                } else if ( player === 'Blue' )
                {
                    blueKills++;
                } else if ( player === '0rigin' )
                {
                    originKills++;
                }
            }
        }
    }

    // Count deaths for tracked players
    if ( splitKill.length > 1 )
    {
        const victimSection = splitKill[ 1 ];
        for ( const player of players )
        {
            if ( victimSection.includes( player ) )
            {
                if ( player === 'Rabbit' )
                {
                    rabbitDeaths++;
                } else if ( player === 'Bear' )
                {
                    bearDeaths++;
                } else if ( player === 'Blue' )
                {
                    blueDeaths++;
                } else if ( player === '0rigin' )
                {
                    originDeaths++;
                }
            }
        }
    }
}
console.log( `Total 'killed' events: ${ count + bearDeaths + blueDeaths + originDeaths + rabbitDeaths }` );
console.log( `Rabbit kills: ${ rabbitKills }` );
console.log( `Bear kills: ${ bearKills }` );
console.log( `Blue kills: ${ blueKills }` );
console.log( `0rigin kills: ${ originKills }` );
console.log( `Rabbit deaths: ${ rabbitDeaths }` );
console.log( `Bear deaths: ${ bearDeaths }` );
console.log( `Blue deaths: ${ blueDeaths }` );
console.log( `0rigin deaths: ${ originDeaths }` );
console.log(`total deaths events: ${ rabbitDeaths + bearDeaths + blueDeaths + originDeaths }`);
console.log( `actual kills: ${ rabbitKills + bearKills + blueKills + originKills }` );