import { describe, it, expect } from 'bun:test';
import { parseLogEvent } from '../src/events';

describe( 'Weapon name mapping', () => {
    it( 'maps blueprint weapon names to clean, user-friendly names', () => {
        const testWeaponLines = [
            '[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_M16A4_C_2147481419',
            '[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_AK74_C_2147481420',
            '[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_M24_C_2147481421',
            '[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_M590A1_C_2147481422',
            '[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Projectile_Molotov_C_2147481423',
            '[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Character_Player_C_2147481424',
            '[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_RPG7_C_2147481425',
            '[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_UnknownWeapon_C_2147481426',
        ];

        const expected = [
            'M16A4',
            'AK-74',
            'M24 SWS',
            'M590A1',
            'Molotov Cocktail',
            'Fall Damage',
            'RPG-7',
            'Unknownweapon',
        ];

        for ( let i = 0; i < testWeaponLines.length; i++ )
        {
            const evt = parseLogEvent( testWeaponLines[ i ] );
            expect( evt ).not.toBeNull();
            if ( !evt ) continue; // satisfy TypeScript narrowing for runtime
            // All provided lines represent player kills
            expect( evt.type ).toMatch( /kill/ );
            expect( evt.data ).toHaveProperty( 'weapon' );
            expect( evt.data.weapon ).toBe( expected[ i ] );
        }
    } );
} );

