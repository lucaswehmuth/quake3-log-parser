# Quake 3 Arena Log Parser

This project is a Go-based log parser for Quake 3 Arena server logs. The parser extracts useful information such as total kills, player scores, and causes of death from the game logs. 

## Features

- Parses log content from Quake 3 Arena server.
- Tracks total kills, including both player and world deaths.
- Updates player scores: players lose 1 point when killed by `<world>` or by suicide.
- Generates a summary report of matches including total kills, player scores, and causes of death.
- Ignores capture the flag events.

## Assumptions

- When `<world>` kills a player, that player loses 1 kill score.
- If a player kills themselves, they lose 1 kill score.
- `<world>` is not considered a player and should not appear in the list of players or in the dictionary of kills.
- The `total_kills` counter includes both player and world deaths.
- The kill counter will not reset if the player disconnects and connects again into the same ongoing match

## Event handling
This code will only consider the below events for parsing:
- InitGame: Initializes a new game.
- ShutdownGame: Ends the current game.
- Kill: Records a kill event, updating the killer's and victim's scores accordingly.
- ClientUserinfoChanged: Tracks changes in player information.

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/lucaswehmuth/quake3-log-parser.git
    cd quake3-log-parser
    ```

2. Install dependencies:
    ```sh
    go mod download
    ```

## Usage

1. Fetch the log file from a URL:
    ```go
    logContent, err := parser.FetchLogFileFromUrl("http://example.com/logfile.log")
    if err != nil {
        log.Fatal(err)
    }
    ```

2. Parse the log content:
    ```go
    lines := strings.Split(logContent, "\n")
    parsedContent := parser.ParseLogContent(lines)
    ```

3. Print the report:
    ```go
    parser.PrintReport(parsedContent.AllMatches)
    ```
    