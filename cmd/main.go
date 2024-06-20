package main

import (
	"fmt"
	"strings"

	"github.com/lucaswehmuth/log-parser/internal/parser"
)

func main() {
	// Example log
	url := "https://gist.githubusercontent.com/cloudwalk-tests/be1b636e58abff14088c8b5309f575d8/raw/df6ef4a9c0b326ce3760233ef24ae8bfa8e33940/qgames.log"

	// Create a parser instance
	logParser := parser.NewQuakeLogParser()

	// Fetch log content
	logContent, err := logParser.FetchLogFileFromUrl(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Parse log content
	logLines := strings.Split(logContent, "\n")
	gameContent := logParser.ParseLogContent(logLines)

	// Print the game matches kill report
	logParser.PrintReport(gameContent.AllMatches)
}
