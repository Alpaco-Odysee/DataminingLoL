package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	apiKey         = "RGAPI-36752775-a120-4542-a136-0410d920e962"
	playerBaseURL  = "https://euw1.api.riotgames.com"
	historyBaseURL = "https://europe.api.riotgames.com"
)

var start = time.Date(2025, 10, 27, 00, 00, 00, 1, time.UTC).Unix()
var end = time.Date(2025, 11, 02, 23, 59, 59, 1, time.UTC).Unix()

type LeagueEntry struct {
	LeagueID string
	PuuID    string
}

type PlayerHistory struct {
	MatchIDs []string
}

func main() {

	createDB()

	// err := handleGetPlayersInRankDivision()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	err := handleGetMatchesFromPlayers()
	if err != nil {
		fmt.Println(err)
	}
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

func getLeagueEntries(pageNumber int) ([]LeagueEntry, error) {
	queue := "RANKED_SOLO_5x5"
	tier := "DIAMOND"
	division := "I"
	url := fmt.Sprintf("%s/lol/league/v4/entries/%s/%s/%s?page=%d", playerBaseURL, queue, tier, division, pageNumber)

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

func handleGetMatchesFromPlayers() error {
	puuIDs, err := getAllPlayerPuuIDs()
	if err != nil {
		return err
	}

	// fmt.Println(puuIDs)
	position := 0
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
		url := fmt.Sprintf("%s/lol/match/v5/matches/by-puuid/%s/ids?startTime=%d&endTime=%d&start=%d&count=%d", historyBaseURL, puuIDs[position], start, end, 0, 100)

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

		var gameHistory []string
		err = json.NewDecoder(resp.Body).Decode(&gameHistory)
		if err != nil {
			return err
		}

		for _, id := range gameHistory {
			fmt.Println(id)
		}

		if len(gameHistory) > 90 {
			fmt.Println("History länger als 90 gefunden")
			break
		} else {
			fmt.Println(position)
			position++
		}
		totalCalls++
	}

	return nil
}
