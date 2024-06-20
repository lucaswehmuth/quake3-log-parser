package parser

import (
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/lucaswehmuth/log-parser/internal/utils"
)

const WORLD_ENTITY = "<world>"

type QuakeParsedContent struct {
	AllMatches []QuakeMatch
}

type QuakeMatch struct {
	TotalKills   int
	PlayerScores map[string]int
	CauseOfDeath map[string]int
}

type QuakeLogParser struct {
	eventHandlers map[string]QuakeEventHandler
	gameRunning   bool
	currentMatch  QuakeMatch
	allMatches    []QuakeMatch
}

// Returns a new QuakeLogParser instance
func NewQuakeLogParser() *QuakeLogParser {
	parser := &QuakeLogParser{
		eventHandlers: make(map[string]QuakeEventHandler),
		gameRunning:   false,
		allMatches:    []QuakeMatch{},
	}
	parser.registerEventHandlers()
	return parser
}

// Private method to register the handlers for each kind of event that needs to be parsed
func (qp *QuakeLogParser) registerEventHandlers() {
	qp.eventHandlers["InitGame"] = &initGameEventHandler{}
	qp.eventHandlers["ShutdownGame"] = &shutdownGameEventHandler{}
	qp.eventHandlers["Kill"] = &killEventHandler{}
	qp.eventHandlers["ClientUserinfoChanged"] = &clientUserinfoChangedEventHandler{}
	// Add more handlers here as needed
}

// FetchLogFileFromUrl fetches data from a given URL and returns the content as a string
func (qp *QuakeLogParser) FetchLogFileFromUrl(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// ParseLogContent parses relevant info (total kills, players kill score, and count of causes of death) about all matches found in the logs
func (qp *QuakeLogParser) ParseLogContent(lines []string) *QuakeParsedContent {
	// Regular expression that matches a relevant quake log event (e.g. '0:00 InitGame:', '20:38 ClientConnect:', '20:54 Kill:')
	eventRe := regexp.MustCompile(`\d+:\d+\s+([^\s:]+):`)

	for _, line := range lines {
		// Check if line contain a valid log format that we are expecting
		keywords := eventRe.FindStringSubmatch(line)
		if len(keywords) <= 1 {
			continue
		}

		// Check if event contains a key of interest for us to parse
		// If so, invoke the corresponding handler to parse that info
		eventKey := keywords[1]
		if handler, exists := qp.eventHandlers[eventKey]; exists {
			handler.HandleEvent(qp, line)
		}
	}

	// If log file ends without a 'ShutdownGame' event we will consider the match as ended
	if qp.gameRunning {
		qp.allMatches = append(qp.allMatches, qp.currentMatch)
		qp.gameRunning = false
	}

	return &QuakeParsedContent{
		AllMatches: qp.allMatches,
	}
}

// PrintReport outputs the scores and stats for each match
func (qp *QuakeLogParser) PrintReport(allMatches []QuakeMatch) {
	fmt.Println("---------------------------------------")
	fmt.Println("Matches kill report:")
	fmt.Println("---------------------------------------")
	for i, game := range allMatches {
		fmt.Printf("Match %d:\n", i+1)
		fmt.Printf("Total Kills: %d\n", game.TotalKills)
		fmt.Println("Scores:")
		for _, entry := range utils.SortMapByValueDescending(game.PlayerScores) {
			fmt.Printf("- %s: %d\n", entry.Key, entry.Value)
		}
		fmt.Println("Cause of Death:")
		for _, entry := range utils.SortMapByValueDescending(game.CauseOfDeath) {
			fmt.Printf("- %s: %d\n", entry.Key, entry.Value)
		}
		fmt.Println("---------------------------------------")
	}
}

// QuakeEventHandler implements an interface for functions to handle parsing of a log event
type QuakeEventHandler interface {
	HandleEvent(parser *QuakeLogParser, line string)
}

// Private handlers for all types of events

type initGameEventHandler struct{}

// HandleEvent handles the logic to parse the start of a new game and eventually end the parsing of a match that was not properly ended.
func (h *initGameEventHandler) HandleEvent(parser *QuakeLogParser, line string) {
	// Based on previous parsed logs the game could apparently quit and show two consecutive 'InitGame' events without a 'ShutdownGame' event (line 98 of example log)
	if parser.gameRunning {
		parser.allMatches = append(parser.allMatches, parser.currentMatch)
	}
	parser.currentMatch = QuakeMatch{
		PlayerScores: make(map[string]int),
		CauseOfDeath: make(map[string]int),
	}
	parser.gameRunning = true
}

type shutdownGameEventHandler struct{}

// HandleEvent handles the logic to finish parsing data for a match that has ended.
func (h *shutdownGameEventHandler) HandleEvent(parser *QuakeLogParser, line string) {
	if parser.gameRunning {
		parser.allMatches = append(parser.allMatches, parser.currentMatch)
		parser.gameRunning = false
	}
}

type killEventHandler struct{}

// HandleEvent handles the logic to parse and update the kill stats of a match.
// The player loses a point if it gets killed by <world>;
// The player loses a point if it kills itself;
// The player wins a point if it kills another player.
func (h *killEventHandler) HandleEvent(parser *QuakeLogParser, line string) {
	// Regular expression that matches a kill event
	// E.g.
	// 2:11 Kill: 2 4 6: Dono da Bola killed Zeh by MOD_ROCKET
	// 21:07 Kill: 1022 2 22: <world> killed Isgalamido by MOD_TRIGGER_HURT
	killRe := regexp.MustCompile(`(\d+:\d+) Kill: \d+ \d+ \d+: (.+) killed (.+) by ([^\s]+)`)
	killMatches := killRe.FindStringSubmatch(line)

	if len(killMatches) > 4 {
		killer := killMatches[2]
		victim := killMatches[3]
		meansOfDeath := killMatches[4]

		if killer != WORLD_ENTITY && killer != victim {
			parser.currentMatch.PlayerScores[killer]++
		} else if killer == WORLD_ENTITY || killer == victim {
			parser.currentMatch.PlayerScores[victim]--
		}

		parser.currentMatch.CauseOfDeath[meansOfDeath]++
		parser.currentMatch.TotalKills++
	}
}

type clientUserinfoChangedEventHandler struct{}

// HandleEvent handles the logic to create a new empty score for users joining a game
func (h *clientUserinfoChangedEventHandler) HandleEvent(parser *QuakeLogParser, line string) {
	// Regular expression that matches a user change event
	// E.g.
	// 3:47 ClientUserinfoChanged: 5 n\Assasinu Credi\t\0\model\sarge\hmodel\sarge\g_redteam\\g_blueteam\\c1\4\c2\5\hc\95\w\0\l\0\tt\0\tl\0
	// 12:14 ClientUserinfoChanged: 3 n\Isgalamido\t\0\model\uriel/zael\hmodel\uriel/zael\g_redteam\\g_blueteam\\c1\5\c2\5\hc\100\w\0\l\0\tt\0\tl\0
	clientRe := regexp.MustCompile(`\d+:\d+ ClientUserinfoChanged: \d+ n\\([^\\]+)\\`)
	clientMatches := clientRe.FindStringSubmatch(line)

	if len(clientMatches) > 1 {
		playerName := clientMatches[1]
		if _, exists := parser.currentMatch.PlayerScores[playerName]; !exists {
			parser.currentMatch.PlayerScores[playerName] = 0
		}
	}
}
