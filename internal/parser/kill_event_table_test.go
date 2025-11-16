// Use package parser_test to allow importing parser functions
package parser_test

import (
	"encoding/json"
	"sandstorm-tracker/internal/parser"
	"testing"
)

type killTestCase struct {
	name    string
	logLine string
	expect  map[string]interface{}
}

func TestParseKillEventTableDriven(t *testing.T) {
	cases := []killTestCase{
		{
			name:    "player killed teammate",
			logLine: `[2025.10.04-15.12.17:473][441]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed Rabbit[76561198995742956, team 0] with BP_Firearm_M16A4_C_2147481419`,
			expect: map[string]interface{}{
				"killers": []map[string]interface{}{{
					"Name":    "ArmoredBear",
					"SteamID": "76561198995742987",
					"Team":    float64(0),
				}},
				"victim": map[string]interface{}{
					"Name":    "Rabbit",
					"SteamID": "76561198995742956",
					"Team":    float64(0),
				},
				"weapon": "M16A4",
			},
		},
		{
			name:    "regular player kills",
			logLine: `[2025.10.04-14.31.05:706][800]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed Marksman[INVALID, team 1] with BP_Firearm_M16A4_C_2147481419`,
			expect: map[string]interface{}{
				"killers": []map[string]interface{}{{
					"Name":    "ArmoredBear",
					"SteamID": "76561198995742987",
					"Team":    float64(0),
				}},
				"victim": map[string]interface{}{
					"Name":    "Marksman",
					"SteamID": "INVALID",
					"Team":    float64(1),
				},
				"weapon": "M16A4",
			},
		},
		{
			name:    "player suicide",
			logLine: `[2025.10.04-15.12.17:473][441]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed ArmoredBear[76561198995742987, team 0] with BP_Character_Player_C_2147481498`,
			expect: map[string]interface{}{
				"killers": []map[string]interface{}{{
					"Name":    "ArmoredBear",
					"SteamID": "76561198995742987",
					"Team":    float64(0),
				}},
				"victim": map[string]interface{}{
					"Name":    "ArmoredBear",
					"SteamID": "76561198995742987",
					"Team":    float64(0),
				},
				"weapon": "Character Player",
			},
		},
		{
			name:    "player suicide molotov",
			logLine: `[2025.10.04-15.22.58:646][535]LogGameplayEvents: Display: ArmoredBear[76561198995742987, team 0] killed ArmoredBear[76561198995742987, team 0] with BP_Projectile_Molotov_C_2147480055`,
			expect: map[string]interface{}{
				"killers": []map[string]interface{}{{
					"Name":    "ArmoredBear",
					"SteamID": "76561198995742987",
					"Team":    float64(0),
				}},
				"victim": map[string]interface{}{
					"Name":    "ArmoredBear",
					"SteamID": "76561198995742987",
					"Team":    float64(0),
				},
				"weapon": "Molotov",
			},
		},
		{
			name:    "multi-player kills",
			logLine: `[2025.10.04-21.29.31:291][459]LogGameplayEvents: Display: -=312th=- Rabbit[76561198262186571, team 0] + *OSS*0rigin[76561198007416544, team 0] killed Rifleman[INVALID, team 1] with BP_Projectile_GAU8_C_2147477120`,
			expect: map[string]interface{}{
				"killers": []map[string]interface{}{
					{
						"Name":    "-=312th=- Rabbit",
						"SteamID": "76561198262186571",
						"Team":    float64(0),
					},
					{
						"Name":    "*OSS*0rigin",
						"SteamID": "76561198007416544",
						"Team":    float64(0),
					},
				},
				"victim": map[string]interface{}{
					"Name":    "Rifleman",
					"SteamID": "INVALID",
					"Team":    float64(1),
				},
				"weapon": "GAU8",
			},
		},
	}

	patterns := parser.NewLogPatterns()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			matches := patterns.PlayerKill.FindStringSubmatch(tc.logLine)
			if len(matches) < 5 {
				t.Fatalf("Kill regex did not match log line: %s", tc.logLine)
			}
			killerSection := matches[2]
			victimSection := matches[3]
			weapon := parser.CleanWeaponName(matches[4])

			killers := parser.ParseKillerSection(killerSection)
			victimParsed := parser.ParseKillerSection(victimSection)
			var victim map[string]interface{}
			if len(victimParsed) > 0 {
				b, _ := json.Marshal(victimParsed[0])
				_ = json.Unmarshal(b, &victim)
			} else {
				victim = map[string]interface{}{"Name": victimSection, "SteamID": "", "Team": -1}
			}

			killersArr := make([]map[string]interface{}, 0, len(killers))
			for _, k := range killers {
				b, _ := json.Marshal(k)
				var m map[string]interface{}
				_ = json.Unmarshal(b, &m)
				killersArr = append(killersArr, m)
			}

			got := map[string]interface{}{
				"killers": killersArr,
				"victim":  victim,
				"weapon":  weapon,
			}

			gotJSON, _ := json.Marshal(got)
			expectJSON, _ := json.Marshal(tc.expect)
			if string(gotJSON) != string(expectJSON) {
				t.Errorf("Parsed event does not match expected.\nGot: %s\nExpected: %s", gotJSON, expectJSON)
			}
		})
	}
}
