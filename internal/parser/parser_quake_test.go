package parser

import (
	"strings"
	"testing"
)

// TestFetchLogFileFromUrl tests the FetchLogFileFromUrl method of QuakeLogParser
func TestFetchLogFileFromUrl(t *testing.T) {
	logParser := NewQuakeLogParser()

	_, err := logParser.FetchLogFileFromUrl("http://example.com/logfile.log")
	if err == nil {
		t.Error("Expected an error for invalid URL, but got none")
	}
}

// TestParseLogContent tests the ParseLogContent method of QuakeLogParser
func TestParseLogContent(t *testing.T) {
	tests := []struct {
		name            string
		logData         string
		expectedMatches []QuakeMatch
	}{
		{
			name: "Single match with one kill",
			logData: `
				11:47 InitGame:
				11:47 ClientUserinfoChanged: 2 n\Dono da Bola\t\0\model\sarge/default\hmodel\sarge/default\g_redteam\g_blueteam\c1\0\c2\0
				11:47 ClientUserinfoChanged: 3 n\Zeh\t\0\model\sarge/default\hmodel\sarge/default\g_redteam\g_blueteam\c1\0\c2\0
				11:47 Kill: 4 2 7: Zeh killed Dono da Bola by MOD_ROCKET_SPLASH
				11:47 ShutdownGame:
			`,
			expectedMatches: []QuakeMatch{
				{
					TotalKills: 1,
					PlayerScores: map[string]int{
						"Zeh":          1,
						"Dono da Bola": 0,
					},
					CauseOfDeath: map[string]int{
						"MOD_ROCKET_SPLASH": 1,
					},
				},
			},
		},
		{
			name: "Multiple matches",
			logData: `
				11:47 InitGame:
				11:47 Kill: 4 2 7: Zeh killed Dono da Bola by MOD_ROCKET_SPLASH
				11:47 ShutdownGame:
				12:00 InitGame:
				12:01 Kill: 4 2 7: Zeh killed Dono da Bola by MOD_ROCKET_SPLASH
				12:01 ShutdownGame:
			`,
			expectedMatches: []QuakeMatch{
				{
					TotalKills: 1,
					PlayerScores: map[string]int{
						"Zeh":          1,
						"Dono da Bola": 0,
					},
					CauseOfDeath: map[string]int{
						"MOD_ROCKET_SPLASH": 1,
					},
				},
				{
					TotalKills: 1,
					PlayerScores: map[string]int{
						"Zeh":          1,
						"Dono da Bola": 0,
					},
					CauseOfDeath: map[string]int{
						"MOD_ROCKET_SPLASH": 1,
					},
				},
			},
		},
		{
			name: "Handle <world> kills",
			logData: `
				11:47 InitGame:
				11:47 Kill: 1022 2 22: <world> killed Dono da Bola by MOD_TRIGGER_HURT
				11:47 ShutdownGame:
			`,
			expectedMatches: []QuakeMatch{
				{
					TotalKills: 1,
					PlayerScores: map[string]int{
						"Dono da Bola": -1,
					},
					CauseOfDeath: map[string]int{
						"MOD_TRIGGER_HURT": 1,
					},
				},
			},
		},
		{
			name: "No kills in match",
			logData: `
				11:47 InitGame:
				11:47 ShutdownGame:
			`,
			expectedMatches: []QuakeMatch{
				{
					TotalKills:   0,
					PlayerScores: map[string]int{},
					CauseOfDeath: map[string]int{},
				},
			},
		},
		{
			name: "Incomplete match (no ShutdownGame)",
			logData: `
				11:47 InitGame:
				11:47 Kill: 4 2 7: Zeh killed Dono da Bola by MOD_ROCKET_SPLASH
			`,
			expectedMatches: []QuakeMatch{
				{
					TotalKills: 1,
					PlayerScores: map[string]int{
						"Zeh":          1,
						"Dono da Bola": 0,
					},
					CauseOfDeath: map[string]int{
						"MOD_ROCKET_SPLASH": 1,
					},
				},
			},
		},
		{
			name: "Two InitGame events without a ShutdownGame",
			logData: `
				20:37 InitGame: \sv_floodProtect\1\sv_maxPing\0\sv_minPing\0\sv_maxRate\10000\sv_minRate\0\sv_hostname\Code Miner Server\g_gametype\0\sv_privateClients\2\sv_maxclients\16\sv_allowDownload\0\bot_minplayers\0\dmflags\0\fraglimit\20\timelimit\15\g_maxGameClients\0\capturelimit\8\version\ioq3 1.36 linux-x86_64 Apr 12 2009\protocol\68\mapname\q3dm17\gamename\baseq3\g_needpass\0
				20:38 ClientConnect: 2
				20:38 ClientUserinfoChanged: 2 n\Isgalamido\t\0\model\uriel/zael\hmodel\uriel/zael\g_redteam\\g_blueteam\\c1\5\c2\5\hc\100\w\0\l\0\tt\0\tl\0
				20:38 ClientBegin: 2
				20:40 Item: 2 weapon_rocketlauncher
				20:40 Item: 2 ammo_rockets
				20:42 Item: 2 item_armor_body
				20:54 Kill: 1022 2 22: <world> killed Isgalamido by MOD_TRIGGER_HURT
				20:59 Item: 2 weapon_rocketlauncher
				21:04 Item: 2 ammo_shells
				21:07 Kill: 1022 2 22: <world> killed Isgalamido by MOD_TRIGGER_HURT
				21:10 ClientDisconnect: 2
				21:15 ClientConnect: 2
				21:15 ClientUserinfoChanged: 2 n\Isgalamido\t\0\model\uriel/zael\hmodel\uriel/zael\g_redteam\\g_blueteam\\c1\5\c2\5\hc\100\w\0\l\0\tt\0\tl\0
				21:17 ClientUserinfoChanged: 2 n\Isgalamido\t\0\model\uriel/zael\hmodel\uriel/zael\g_redteam\\g_blueteam\\c1\5\c2\5\hc\100\w\0\l\0\tt\0\tl\0
				21:17 ClientBegin: 2
				21:42 Kill: 1022 2 22: <world> killed Isgalamido by MOD_TRIGGER_HURT
				21:49 Item: 2 weapon_rocketlauncher
				21:51 ClientConnect: 3
				21:51 ClientUserinfoChanged: 3 n\Dono da Bola\t\0\model\sarge/krusade\hmodel\sarge/krusade\g_redteam\\g_blueteam\\c1\5\c2\5\hc\95\w\0\l\0\tt\0\tl\0
				21:53 ClientUserinfoChanged: 3 n\Mocinha\t\0\model\sarge\hmodel\sarge\g_redteam\\g_blueteam\\c1\4\c2\5\hc\95\w\0\l\0\tt\0\tl\0
				21:53 ClientBegin: 3
				22:04 Item: 2 weapon_rocketlauncher
				22:04 Item: 2 ammo_rockets
				22:06 Kill: 2 3 7: Isgalamido killed Mocinha by MOD_ROCKET_SPLASH
				22:11 Item: 2 item_quad
				22:11 ClientDisconnect: 3
				22:18 Kill: 2 2 7: Isgalamido killed Isgalamido by MOD_ROCKET_SPLASH
				22:26 Item: 2 weapon_rocketlauncher
				22:27 Item: 2 ammo_rockets
				22:40 Kill: 2 2 7: Isgalamido killed Isgalamido by MOD_ROCKET_SPLASH
				22:43 Item: 2 weapon_rocketlauncher
				22:45 Item: 2 item_armor_body
				23:06 Kill: 1022 2 22: <world> killed Isgalamido by MOD_TRIGGER_HURT
				25:05 Kill: 1022 2 22: <world> killed Isgalamido by MOD_TRIGGER_HURT
				25:09 Item: 2 weapon_rocketlauncher
				25:09 Item: 2 ammo_rockets
				25:11 Item: 2 item_armor_body
				25:18 Kill: 1022 2 22: <world> killed Isgalamido by MOD_TRIGGER_HURT
				25:21 Item: 2 weapon_rocketlauncher
				25:22 Item: 2 ammo_rockets
				25:34 Item: 2 weapon_rocketlauncher
				25:41 Kill: 1022 2 19: <world> killed Isgalamido by MOD_FALLING
				25:50 Item: 2 item_armor_combat
				25:52 Kill: 1022 2 22: <world> killed Isgalamido by MOD_TRIGGER_HURT
				26  0:00 ------------------------------------------------------------
				0:00 InitGame: \sv_floodProtect\1\sv_maxPing\0\sv_minPing\0\sv_maxRate\10000\sv_minRate\0\sv_hostname\Code Miner Server\g_gametype\0\sv_privateClients\2\sv_maxclients\16\sv_allowDownload\0\dmflags\0\fraglimit\20\timelimit\15\g_maxGameClients\0\capturelimit\8\version\ioq3 1.36 linux-x86_64 Apr 12 2009\protocol\68\mapname\q3dm17\gamename\baseq3\g_needpass\0
				0:25 ClientConnect: 2
				0:25 ClientUserinfoChanged: 2 n\Dono da Bola\t\0\model\sarge/krusade\hmodel\sarge/krusade\g_redteam\\g_blueteam\\c1\5\c2\5\hc\95\w\0\l\0\tt\0\tl\0
				0:27 ClientUserinfoChanged: 2 n\Mocinha\t\0\model\sarge\hmodel\sarge\g_redteam\\g_blueteam\\c1\4\c2\5\hc\95\w\0\l\0\tt\0\tl\0
				0:27 ClientBegin: 2
				0:29 Item: 2 weapon_rocketlauncher
				0:55 Item: 2 item_health_large
				0:56 Item: 2 weapon_rocketlauncher
				0:57 Item: 2 ammo_rockets
				0:59 ClientConnect: 3
				0:59 ClientUserinfoChanged: 3 n\Isgalamido\t\0\model\xian/default\hmodel\xian/default\g_redteam\\g_blueteam\\c1\4\c2\5\hc\100\w\0\l\0\tt\0\tl\0
				1:01 ClientUserinfoChanged: 3 n\Isgalamido\t\0\model\uriel/zael\hmodel\uriel/zael\g_redteam\\g_blueteam\\c1\5\c2\5\hc\100\w\0\l\0\tt\0\tl\0
				1:01 ClientBegin: 3
				1:02 Item: 3 weapon_rocketlauncher
				1:06 ClientConnect: 4
				1:06 ClientUserinfoChanged: 4 n\Zeh\t\0\model\sarge/default\hmodel\sarge/default\g_redteam\\g_blueteam\\c1\5\c2\5\hc\100\w\0\l\0\tt\0\tl\0
				1:08 Kill: 3 2 6: Isgalamido killed Mocinha by MOD_ROCKET
				1:08 ClientUserinfoChanged: 4 n\Zeh\t\0\model\sarge/default\hmodel\sarge/default\g_redteam\\g_blueteam\\c1\1\c2\5\hc\100\w\0\l\0\tt\0\tl\0
				1:08 ClientBegin: 4
				1:18 Item: 4 weapon_rocketlauncher
				1:26 Kill: 1022 4 22: <world> killed Zeh by MOD_TRIGGER_HURT
				1:26 ClientUserinfoChanged: 2 n\Dono da Bola\t\0\model\sarge\hmodel\sarge\g_redteam\\g_blueteam\\c1\4\c2\5\hc\95\w\0\l\0\tt\0\tl\0
				1:26 Item: 3 weapon_railgun
				1:29 Item: 2 weapon_rocketlauncher
				1:32 Kill: 1022 4 22: <world> killed Zeh by MOD_TRIGGER_HURT
				1:35 Item: 3 weapon_railgun
				1:38 Item: 3 weapon_railgun
				1:41 Kill: 1022 2 19: <world> killed Dono da Bola by MOD_FALLING
				1:41 Item: 3 weapon_railgun
				1:44 Item: 2 weapon_rocketlauncher
				1:47 ShutdownGame:
			`,
			expectedMatches: []QuakeMatch{
				{
					TotalKills: 11,
					PlayerScores: map[string]int{
						"Dono da Bola": 0,
						"Mocinha":      0,
						"Isgalamido":   -9,
					},
					CauseOfDeath: map[string]int{
						"MOD_TRIGGER_HURT":  7,
						"MOD_ROCKET_SPLASH": 3,
						"MOD_FALLING":       1,
					},
				},
				{
					TotalKills: 4,
					PlayerScores: map[string]int{
						"Isgalamido":   1,
						"Mocinha":      0,
						"Dono da Bola": -1,
						"Zeh":          -2,
					},
					CauseOfDeath: map[string]int{
						"MOD_TRIGGER_HURT": 2,
						"MOD_ROCKET":       1,
						"MOD_FALLING":      1,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := strings.Split(tt.logData, "\n")
			quakeParser := NewQuakeLogParser()
			parsedContent := quakeParser.ParseLogContent(lines)

			if len(parsedContent.AllMatches) != len(tt.expectedMatches) {
				t.Fatalf("Expected %d matches, but got %d", len(tt.expectedMatches), len(parsedContent.AllMatches))
			}

			for i, match := range parsedContent.AllMatches {
				if match.TotalKills != tt.expectedMatches[i].TotalKills {
					t.Errorf("Expected %d kills, but got %d", tt.expectedMatches[i].TotalKills, match.TotalKills)
				}

				for player, score := range tt.expectedMatches[i].PlayerScores {
					if match.PlayerScores[player] != score {
						t.Errorf("Expected %d kills for player %s, but got %d", score, player, match.PlayerScores[player])
					}
				}

				for cause, count := range tt.expectedMatches[i].CauseOfDeath {
					if match.CauseOfDeath[cause] != count {
						t.Errorf("Expected %d deaths by %s, but got %d", count, cause, match.CauseOfDeath[cause])
					}
				}
			}
		})
	}
}
