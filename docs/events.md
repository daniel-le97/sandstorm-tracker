// game over
[2025.10.04-15.23.38:790][979]LogGameplayEvents: Display: Game over

// player joined
[2025.10.04-15.35.33:666][194]LogNet: Join succeeded: ArmoredBear


// player left
[2025.10.04-15.37.59:204][779]LogRcon: 127.0.0.1:58877 << say See you later, ArmoredBear!

// player disconnected / left
[2025.10.04-13.50.55:457][944]LogEOSAntiCheat: Display: ServerUnregisterClient: UserId (76561198995742987), Result: (EOS_Success)

// player killed team mate
[2025.10.04-15.12.17:473][441]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed Rabbit[76561198995742956, team 0] with BP_Firearm_M16A4_C_2147481419

// player kills
[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed Marksman[INVALID, team 1] with BP_Firearm_M16A4_C_2147481419
[2025.10.04-14.31.10:921][111]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed Observer[INVALID, team 1] with BP_Firearm_M16A4_C_2147481419
[2025.10.04-14.31.29:670][226]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed Commander[INVALID, team 1] with BP_Firearm_M16A4_C_2147481419

// ai bot killed itself
[2025.10.04-19.51.47:382][508]LogGameplayEvents: Display: ? killed Observer[INVALID, team 1] with BP_Projectile_Mortar_HE_C_2147480348


// difficulty
[2025.10.04-15.34.56:470][  0]LogAI: Warning: AI difficulty set to 0.5

// rounds
[2025.10.04-15.21.46:114][183]LogGameMode: Display: Round O ver: Team 1 won (win reason: Elimination)
[2025.10.04-15.21.46:114][183]LogGameplayEvents: Display: Round 2 Over: Team 1 won (win reason: Elimination)

// map vote
[2025.10.04-15.23.38:812][979]LogMapVoteManager: Display: Existing Vote Options:
[2025.10.04-15.23.38:812][979]LogMapVoteManager: Display: New Vote Options:

// player died to fall / suicide
[2025.10.04-15.12.17:472][441]LogSoldier: Applying 268.43 fall damage, downward velocity on landing was -1821.08
[2025.10.04-15.12.17:473][441]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed ArmoredBear[76561198995742987, team 0] with BP_Character_Player_C_2147481498

// suicide
[2025.10.04-15.22.58:646][535]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed ArmoredBear[76561198995742987, team 0] with BP_Projectile_Molotov_C_2147480055

// loading map - map:crossing - team:security
[2025.10.04-13.46.26:141][  0]LogLoad: LoadMap: /Game/Maps/Canyon/Canyon?Name=Player?Scenario=Scenario_Crossing_Checkpoint_Security?MaxPlayers=8?Lighting=Day


// player requesting their stats, this shows score/min and their rank out of all players on the server
// at the moment im not sure how to get the rank, but we can calculate it from total kills for now
[2025.10.04-16.42.23:199][613]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !stats

// player requesting another players stats - should show kills, deaths, kdr
[2025.10.04-16.42.23:199][613]LogChat: Display: Rabbit(76561198995742956) Global Chat: !stats Armoredbear

// player requesting another players stats - should show kills, deaths, kdr - if player is not found, we should find the most likely match. In this case ArmoredBear's stats should be shown, if there are multiple matches we should show the closest match. if this can not find anyone, we should say player(s) not found
[2025.10.04-16.42.23:199][613]LogChat: Display: Rabbit(76561198995742956) Global Chat: !stats Armo

// player requesting their kdr - should show kills, deaths, kdr
[2025.10.04-16.42.26:896][833]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !kdr

// player requesting top 3 players - should show the top 3 players on the server by kills
[2025.10.04-16.42.26:896][833]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !top

// player requesting weapons - should show the players top 3 used weapons and their kills with each weapon
[2025.10.04-16.42.31:683][118]LogChat: Display: ArmoredBear(76561198995742987) Global Chat: !guns