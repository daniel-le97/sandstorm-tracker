import { parseLogEvent } from '../events';

export function runWeaponTests (): void {
    console.log( '🔫 Testing Weapon Name Mappings:\n' );

    try
    {
        const testWeaponLines = [
            "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_M16A4_C_2147481419",
            "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_AK74_C_2147481420",
            "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_M24_C_2147481421",
            "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_M590A1_C_2147481422",
            "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Projectile_Molotov_C_2147481423",
            "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Character_Player_C_2147481424",
            "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_RPG7_C_2147481425",
            "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_UnknownWeapon_C_2147481426"
        ];

        testWeaponLines.forEach( ( line, index ) => {
            const event = parseLogEvent( line );
            if ( event && event.type.includes( 'kill' ) )
            {
                console.log( `${ index + 1 }. Original: ${ line.match( /with (.+)$/ )?.[ 1 ] }` );
                console.log( `   Clean:    ${ event.data.weapon }\n` );
            }
        } );

        console.log( '✅ Weapon tests completed successfully!' );
    } catch ( error )
    {
        console.error( '\n❌ Weapon tests failed:', error );
        throw error;
    }
}