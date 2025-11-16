package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	apiKey         = "RGAPI-36ae3905-c7da-427e-9ad0-78848857456b"
	playerBaseURL  = "https://euw1.api.riotgames.com"
	historyBaseURL = "https://europe.api.riotgames.com"
)

var start = time.Date(2025, 11, 16, 00, 00, 00, 1, time.UTC).Unix()
var end = time.Date(2025, 11, 16, 23, 59, 59, 1, time.UTC).Unix()

func main() {
	_, err := os.Stat("../app.log")

	var file *os.File
	if os.IsNotExist(err) {
		file, err = os.Create("../app.log")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		file, err = os.OpenFile("../app.log", os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer file.Close()

	log.SetOutput(file)

	log.Println("App started")
	err = handleGetPlayersInRankDivision()
	if err != nil {
		log.Println("Error:", err)
	} else {
		log.Println("Successfully saved JSON response to players.json")
	}
	err = handleGetMatchesFromPlayers()
	if err != nil {
		log.Println("Error:", err)
	} else {
		log.Println("Successfully saved JSON response to matches.json")
	}

	log.Println("App finished")
}

func handleGetPlayersInRankDivision() error {
	queue := "RANKED_SOLO_5x5"
	tier := "DIAMOND"
	division := "I"
	url := fmt.Sprintf("%s/lol/league/v4/entries/%s/%s/%s?page=%d", playerBaseURL, queue, tier, division, 1)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Riot-Token", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var data interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return err
	}

	dateDir := time.Now().Format("2006-01-02")
	saveDir := filepath.Join("../logs", dateDir)
	err = os.MkdirAll(saveDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	outPath := filepath.Join(saveDir, "players.json")
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	log.Printf("Saved player data to %s\n", outPath)
	return nil
}

func handleGetMatchesFromPlayers() error {
	log.Println("Starting handleGetMatchesFromPlayers...")

	dateDir := time.Now().Format("2006-01-02")
	saveDir := filepath.Join("../logs", dateDir)
	os.MkdirAll(saveDir, 0755)

	playersPath := filepath.Join(saveDir, "players.json")

	file, err := os.Open(playersPath)
	if err != nil {
		return fmt.Errorf("could not open players.json: %w", err)
	}
	defer file.Close()

	var players []map[string]interface{}
	if err := json.NewDecoder(file).Decode(&players); err != nil {
		return fmt.Errorf("failed to decode players.json: %w", err)
	}

	if len(players) == 0 {
		return fmt.Errorf("players.json is empty")
	}

	log.Printf("Loaded %d players\n", len(players))

	client := &http.Client{}
	totalCalls := 0

	for i := range players {
		if totalCalls%20 == 0 {
			log.Println("waiting 2 seconds")
			time.Sleep(2 * time.Second)
		}
		if totalCalls == 100 {
			log.Println("waiting 241 seconds")
			time.Sleep(241 * time.Second)
			totalCalls = 0
		}

		puuid, ok := players[i]["puuid"].(string)
		if !ok || puuid == "" {
			log.Printf("Skipping player %d: no puuid\n", i)
			continue
		}

		log.Printf("Checking PUUID #%d: %s\n", i, puuid)

		matchListURL := fmt.Sprintf(
			"%s/lol/match/v5/matches/by-puuid/%s/ids?startTime=%d&endTime=%d&start=0&count=20",
			historyBaseURL, puuid, start, end,
		)

		req, err := http.NewRequest("GET", matchListURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("X-Riot-Token", apiKey)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		var matchIDs []string
		if err := json.NewDecoder(resp.Body).Decode(&matchIDs); err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed to decode match list: %w", err)
		}
		resp.Body.Close()
		totalCalls++

		matchListPath := filepath.Join(saveDir, "matchList.json")
		matchListFile, err := os.Create(matchListPath)
		if err != nil {
			return fmt.Errorf("failed to create matchList.json: %w", err)
		}

		mlEnc := json.NewEncoder(matchListFile)
		mlEnc.SetIndent("", "  ")
		if err := mlEnc.Encode(matchIDs); err != nil {
			matchListFile.Close()
			return fmt.Errorf("failed to write match list: %w", err)
		}
		matchListFile.Close()

		log.Printf("Saved match list (%d entries) to %s\n", len(matchIDs), matchListPath)

		if len(matchIDs) == 0 {
			log.Printf("PUUID %s has no matches. Skipping...\n", puuid)
			continue
		}

		matchID := matchIDs[0]
		log.Printf("Fetching match %s\n", matchID)

		matchURL := fmt.Sprintf("%s/lol/match/v5/matches/%s", historyBaseURL, matchID)

		req2, err := http.NewRequest("GET", matchURL, nil)
		if err != nil {
			return err
		}
		req2.Header.Set("X-Riot-Token", apiKey)

		resp2, err := client.Do(req2)
		if err != nil {
			return err
		}

		var matchData interface{}
		if err := json.NewDecoder(resp2.Body).Decode(&matchData); err != nil {
			resp2.Body.Close()
			return fmt.Errorf("failed to decode match data: %w", err)
		}
		resp2.Body.Close()

		matchPath := filepath.Join(saveDir, "match.json")
		matchFile, err := os.Create(matchPath)
		if err != nil {
			return fmt.Errorf("failed to create match.json: %w", err)
		}

		mEnc := json.NewEncoder(matchFile)
		mEnc.SetIndent("", "  ")
		if err := mEnc.Encode(matchData); err != nil {
			matchFile.Close()
			return fmt.Errorf("failed to save match data: %w", err)
		}
		matchFile.Close()

		log.Printf("Saved match data to %s\n", matchPath)
		return nil
	}

	return fmt.Errorf("no matches found for any puuid")
}
