package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiKey = "RGAPI-a83dbc55-39d2-4521-96f7-3ead65ea4d3d"
const baseURL = "https://euw1.api.riotgames.com"

type LeagueEntry struct {
	LeagueID string
	PuuID    string
}

func main() {
	createDB()

	err := handleGetPlayersInRankDivision()
	if err != nil {
		fmt.Println(err)
	}
	// err = handleGetMatchesFromPlayers()
	// if err != nil {
	// 	fmt.Println(err)
	// }
}

func handleGetPlayersInRankDivision() error {
	page := 1
	totalCalls := 0

	for {
		if totalCalls%20 == 0 {
			fmt.Println("waiting 2 seconds")
			time.Sleep(2 * time.Second)
		}
		if totalCalls == 100 {
			fmt.Println("waiting 241 seconds")
			time.Sleep(241 * time.Second)
			totalCalls = 0
		}

		entries, err := getLeagueEntries(page)
		if err != nil {
			return err
		}

		if len(entries) == 0 {
			fmt.Printf("Reached end of pages at page %d.\n", page)
			break
		}

		fmt.Printf("Page %d - %d entries\n", page, len(entries))

		for _, entry := range entries {
			addPlayer(entry)
			fmt.Printf("League ID: %s | PuuID: %s\n", entry.LeagueID, entry.PuuID)
		}

		page++
		totalCalls++
	}

	return nil
}

// func handleGetMatchesFromPlayers() error {
// 	puuIDs, err := getAllPlayerPuuIDs()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func getLeagueEntries(pageNumber int) ([]LeagueEntry, error) {
	queue := "RANKED_SOLO_5x5"
	tier := "DIAMOND"
	division := "I"
	url := fmt.Sprintf("%s/lol/league/v4/entries/%s/%s/%s?page=%d", baseURL, queue, tier, division, pageNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Riot-Token", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var entries []LeagueEntry
	err = json.NewDecoder(resp.Body).Decode(&entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}
