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
	apiKey         = "RGAPI-e0def929-8101-43a4-96f6-6fe9ba56edaf"
	playerBaseURL  = "https://euw1.api.riotgames.com"
	historyBaseURL = "https://europe.api.riotgames.com"
)

var start = time.Date(2025, 11, 16, 00, 00, 00, 1, time.UTC).Unix()
var end = time.Date(2025, 11, 16, 23, 59, 59, 1, time.UTC).Unix()

func main() {
	// Check if the file exists
	_, err := os.Stat("app.log")

	var file *os.File
	if os.IsNotExist(err) {
		// Create file if it doesn't exist
		file, err = os.Create("app.log")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Open file for appending if it exists
		file, err = os.OpenFile("app.log", os.O_APPEND|os.O_WRONLY, 0666)
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
	saveDir := filepath.Join("logs", dateDir)
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
