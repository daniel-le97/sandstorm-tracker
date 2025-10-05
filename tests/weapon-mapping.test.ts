import { test, expect, describe } from 'bun:test';
import { parseLogEvent } from '../src/events';

describe( 'Weapon Name Mapping', () => {
    const weaponTestCases = [
        {
            blueprint: 'BP_Firearm_M16A4_C_2147481419',
            expected: 'M16A4',
            description: 'M16A4 rifle'
        },
        {
            blueprint: 'BP_Firearm_AK74_C_2147481420',
            expected: 'AK-74',
            description: 'AK-74 rifle'
        },
        {
            blueprint: 'BP_Firearm_M24_C_2147481421',
            expected: 'M24 SWS',
            description: 'M24 sniper rifle'
        },
        {
            blueprint: 'BP_Firearm_M590A1_C_2147481422',
            expected: 'M590A1',
            description: 'M590A1 shotgun'
        },
        {
            blueprint: 'BP_Projectile_Molotov_C_2147481423',
            expected: 'Molotov Cocktail',
            description: 'Molotov cocktail'
        },
        {
            blueprint: 'BP_Character_Player_C_2147481424',
            expected: 'Fall Damage',
            description: 'Fall damage'
        },
        {
            blueprint: 'BP_Firearm_RPG7_C_2147481425',
            expected: 'RPG-7',
            description: 'RPG-7 launcher'
        },
        {
            blueprint: 'BP_Firearm_UnknownWeapon_C_2147481426',
            expected: 'Unknownweapon',
            description: 'Unknown weapon fallback'
        }
    ];

    test.each( weaponTestCases )( '$description maps correctly', ( { blueprint, expected } ) => {
        const logLine = `[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with ${ blueprint }`;
        const event = parseLogEvent( logLine );

        expect( event ).not.toBeNull();
        expect( event?.type ).toMatch( /kill/ );
        expect( event?.data.weapon ).toBe( expected );
    } );

    test( 'Weapon name mapping preserves original if no mapping exists', () => {
        const blueprint = 'BP_Firearm_CustomWeapon_C_123456';
        const logLine = `[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with ${ blueprint }`;
        const event = parseLogEvent( logLine );

        expect( event ).not.toBeNull();
        expect( event?.data.weapon ).toBe( 'Customweapon' ); // Should fall back to cleaned blueprint name
    } );

    test( 'All weapon mappings show user-friendly names', () => {
        const testWeaponLines = [
            "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_M16A4_C_2147481419",
            "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_AK74_C_2147481420",
            "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_M24_C_2147481421",
            "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_M590A1_C_2147481422",
            "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Projectile_Molotov_C_2147481423",
            "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Character_Player_C_2147481424",
            "[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: Player1[12345, team 0] killed Player2[67890, team 1] with BP_Firearm_RPG7_C_2147481425"
        ];

        const expectedWeapons = [ 'M16A4', 'AK-74', 'M24 SWS', 'M590A1', 'Molotov Cocktail', 'Fall Damage', 'RPG-7' ];

        testWeaponLines.forEach( ( line, index ) => {
            const event = parseLogEvent( line );
            expect( event ).not.toBeNull();
            expect( event?.data.weapon ).toBe( expectedWeapons[ index ] );
            expect( event?.data.weapon ).not.toMatch( /BP_/ ); // Should not contain blueprint prefix
        } );
    } );
} );