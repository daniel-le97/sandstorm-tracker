import { beforeAll, describe, expect, test } from "bun:test";
import { parseLogEvent } from "../src/events";

describe( "Commander Fire Support Integration", () => {
    let TrackerService: any;
    let testServerDbId: number;
    let db: any;

    beforeAll( async () => {
        // Set test environment
        process.env.TEST_DB_PATH = "test_fire_support.db";

        // Import modules
        const dbModule = await import( "../src/database.ts" );
        TrackerService = ( await import( "../src/trackerService.ts" ) ).default;
        db = dbModule.default();

        // Clean up any existing data
        db.run( 'DELETE FROM kills' );
        db.run( 'DELETE FROM player_sessions' );
        db.run( 'DELETE FROM players' );
        db.run( 'DELETE FROM map_rounds' );
        db.run( 'DELETE FROM maps' );
        db.run( 'DELETE FROM servers' );

        // Generate unique config ID to avoid UNIQUE constraint failures
        const uniqueId = Date.now().toString();

        // Create test server first (required for foreign key constraints)
        testServerDbId = dbModule.upsertServer(
            `test-server-fire-support-${ uniqueId }`,
            "Test Server Fire Support",
            `test-config-fire-support-${ uniqueId }`,
            "/test/logs/test-server-fire-support.log",
            "Test server for fire support integration tests"
        );
    } );

    test( "Fire support weapon mapping works correctly", () => {
        const fireSuportWeaponLines = [
            "[2025.10.05-19.51.47:382][508]LogGameplayEvents: Display: ? killed Observer[INVALID, team 1] with BP_Projectile_Mortar_HE_C_2147480348",
            "[2025.10.05-19.51.50:382][508]LogGameplayEvents: Display: Commander[76561198123456789, team 0] killed Enemy1[76561198987654321, team 1] with BP_Projectile_Artillery_HE_C_2147480349",
            "[2025.10.05-19.51.55:382][508]LogGameplayEvents: Display: ? killed Enemy2[INVALID, team 1] with BP_Vehicle_Helicopter_Gunship_C_2147480350",
            "[2025.10.05-19.52.00:382][508]LogGameplayEvents: Display: Commander[76561198123456789, team 0] killed Enemy3[76561198987654322, team 1] with BP_Projectile_Hellfire_C_2147480351",
            "[2025.10.05-19.52.05:382][508]LogGameplayEvents: Display: ? killed Enemy4[INVALID, team 1] with BP_Projectile_Airstrike_C_2147480352"
        ];

        const expectedWeapons = [
            'Mortar Strike',
            'Artillery Strike',
            'Attack Helicopter',
            'Hellfire Missile',
            'Air Strike'
        ];

        fireSuportWeaponLines.forEach( ( line, index ) => {
            const event = parseLogEvent( line );
            expect( event ).not.toBeNull();
            expect( event?.type ).toMatch( /player_kill|suicide/ );
            expect( event?.data.weapon ).toBe( expectedWeapons[ index ] );
            expect( event?.data.weapon ).not.toMatch( /BP_/ ); // Should not contain blueprint prefix
        } );
    } );

    test( "Fire support kills are properly recorded in database", () => {
        // Process mortar kill event
        const mortarEvent = parseLogEvent(
            "[2025.10.05-19.51.47:382][508]LogGameplayEvents: Display: ? killed TestVictim[76561198999888777, team 1] with BP_Projectile_Mortar_HE_C_2147480348"
        );

        expect( mortarEvent ).not.toBeNull();
        if ( mortarEvent )
        {
            // Process the event through stats service
            TrackerService.processEvent( mortarEvent, testServerDbId );

            // Verify the kill was recorded (even with unknown killer)
            // Note: The system may create a placeholder entry for unknown killers
        }

        // Process artillery kill with known commander
        const artilleryEvent = parseLogEvent(
            "[2025.10.05-19.52.00:382][508]LogGameplayEvents: Display: CommanderPlayer[76561198123456789, team 0] killed ArtilleryVictim[76561198987654321, team 1] with BP_Projectile_Artillery_HE_C_2147480349"
        );

        expect( artilleryEvent ).not.toBeNull();
        if ( artilleryEvent )
        {
            TrackerService.processEvent( artilleryEvent, testServerDbId );

            // Check that commander player stats include the artillery kill
            const commanderStats = TrackerService.getPlayerStats( "76561198123456789", testServerDbId );
            expect( commanderStats ).toBeDefined();
            if ( commanderStats )
            {
                expect( commanderStats.total_kills ).toBeGreaterThan( 0 );
            }

            // Check weapon stats for artillery
            const weaponStats = TrackerService.getPlayerWeapons( "76561198123456789", testServerDbId, 10 );
            expect( weaponStats ).toBeDefined();

            const artilleryWeaponStat = weaponStats?.find( ( w: any ) => w.weapon_name === "Artillery Strike" );
            expect( artilleryWeaponStat ).toBeDefined();
            if ( artilleryWeaponStat )
            {
                expect( artilleryWeaponStat.kills ).toBeGreaterThan( 0 );
            }
        }
    } );

    test( "All fire support weapons have proper mappings", () => {
        const fireSupportBlueprintNames = [
            "BP_Projectile_Mortar_HE_C_123",
            "BP_Projectile_Mortar_Smoke_C_456",
            "BP_Projectile_Artillery_HE_C_789",
            "BP_Projectile_Artillery_Smoke_C_012",
            "BP_Vehicle_Helicopter_Gunship_C_345",
            "BP_Projectile_Hellfire_C_678",
            "BP_Projectile_Airstrike_C_901",
            "BP_Projectile_Strafe_C_234",
            "BP_Projectile_Rocket_155mm_C_567",
            "BP_Projectile_Rocket_120mm_C_890"
        ];

        const expectedNames = [
            "Mortar Strike",
            "Mortar Smoke",
            "Artillery Strike",
            "Artillery Smoke",
            "Attack Helicopter",
            "Hellfire Missile",
            "Air Strike",
            "Strafing Run",
            "155mm Artillery",
            "120mm Mortar"
        ];

        // Import the weapon mapping function
        const events = require( "../src/events" );
        const getCleanWeaponName = events.getCleanWeaponName ||
            ( ( name: string ) => events.parseLogEvent( `[test] LogGameplayEvents: Display: Test[123, team 0] killed Test2[456, team 1] with ${ name }` )?.data?.weapon );

        fireSupportBlueprintNames.forEach( ( blueprintName, index ) => {
            // Create a test kill event to process the weapon name
            const testLine = `[2025.10.05-12.00.00:000][123]LogGameplayEvents: Display: TestKiller[76561198000000001, team 0] killed TestVictim[76561198000000002, team 1] with ${ blueprintName }`;
            const event = parseLogEvent( testLine );

            expect( event ).not.toBeNull();
            expect( event?.data.weapon ).toBe( expectedNames[ index ] );
        } );
    } );

    test( "Fire support kills show up in weapon statistics", () => {
        // Create test player and add some fire support kills
        const testEvents = [
            "[2025.10.05-12.01.00:000][123]LogGameplayEvents: Display: FireSupportCommander[76561198111222333, team 0] killed Enemy1[76561198444555666, team 1] with BP_Projectile_Mortar_HE_C_123",
            "[2025.10.05-12.02.00:000][124]LogGameplayEvents: Display: FireSupportCommander[76561198111222333, team 0] killed Enemy2[76561198444555667, team 1] with BP_Projectile_Artillery_HE_C_456",
            "[2025.10.05-12.03.00:000][125]LogGameplayEvents: Display: FireSupportCommander[76561198111222333, team 0] killed Enemy3[76561198444555668, team 1] with BP_Projectile_Mortar_HE_C_789"
        ];

        testEvents.forEach( line => {
            const event = parseLogEvent( line );
            if ( event )
            {
                TrackerService.processEvent( event, testServerDbId );
            }
        } );

        // Check weapon stats for the commander
        const weaponStats = TrackerService.getPlayerWeapons( "76561198111222333", testServerDbId, 10 );
        expect( weaponStats ).toBeDefined();
        expect( weaponStats?.length ).toBeGreaterThan( 0 );

        // Check that fire support weapons appear in stats
        const mortarStats = weaponStats?.find( ( w: any ) => w.weapon_name === "Mortar Strike" );
        const artilleryStats = weaponStats?.find( ( w: any ) => w.weapon_name === "Artillery Strike" );

        expect( mortarStats ).toBeDefined();
        expect( artilleryStats ).toBeDefined();

        if ( mortarStats )
        {
            expect( mortarStats.kills ).toBe( 2 ); // Two mortar kills
        }
        if ( artilleryStats )
        {
            expect( artilleryStats.kills ).toBe( 1 ); // One artillery kill
        }
    } );
} );
