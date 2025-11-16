package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"
)

const (
	apiKey         = "RGAPI-d7859f0d-11c7-4fa7-b720-909190c75eab"
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
	// err := handleGetMatchesFromPlayers()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	err := handleGetTargetData()
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

type GameHistory struct {
	PuuID           string   `json: "puuid"`
	MatchhistoryIDs []string `json: "history"`
}

type FirstBlood struct {
	Champion string
	Win      int
	MaxGames int
}

func handleGetMatchesFromPlayers() error {
	puuIDs, err := getAllPlayerPuuIDs()
	if err != nil {
		return err
	}

	totalCalls := 0
	for _, puuID := range puuIDs {
		if totalCalls%20 == 0 {
			fmt.Println("waiting 2 seconds")
			time.Sleep(2 * time.Second)
		}
		if totalCalls == 100 {
			fmt.Println("waiting 241 seconds")
			time.Sleep(241 * time.Second)
			totalCalls = 0
		}
		url := fmt.Sprintf("%s/lol/match/v5/matches/by-puuid/%s/ids?startTime=%d&endTime=%d&start=%d&count=%d", historyBaseURL, puuID, start, end, 0, 100)

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

		gHistory := GameHistory{}
		gHistory.PuuID = puuID
		for _, id := range gameHistory {
			gHistory.MatchhistoryIDs = append(gHistory.MatchhistoryIDs, id)
			fmt.Println(id)
		}

		err = addPlayerHistory(gHistory)
		if err != nil {
			return err
		}
		totalCalls++
	}
	return nil
}

func handleGetTargetData() error {
	histories, err := getMatchHistoryFromPlayer()
	if err != nil {
		return err
	}
	var exists []string
	firstBloodStats := make(map[string]*FirstBlood)

	totalCalls := 0
	for historyMax, h := range histories {
		fmt.Printf("PUUID: %s | Matches: %d\n", h.PuuID, len(h.MatchhistoryIDs))
		for _, match := range h.MatchhistoryIDs {
			if totalCalls%20 == 0 {
				fmt.Println("waiting 2 seconds")
				time.Sleep(2 * time.Second)
			}
			if totalCalls == 100 {
				fmt.Println("waiting 241 seconds")
				time.Sleep(241 * time.Second)
				totalCalls = 0
			}
			if slices.Contains(exists, match) {
				continue
			}
			url := fmt.Sprintf("%s/lol/match/v5/matches/%s", historyBaseURL, match)

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

			var matchData map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&matchData); err != nil {
				return err
			}
			info, ok := matchData["info"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("unexpected match structure")
			}
			participants, ok := info["participants"].([]interface{})
			if !ok {
				return fmt.Errorf("no participants found")
			}

			for i := range len(participants) {
				first := participants[i].(map[string]interface{})

				champion, _ := first["championName"].(string)
				win, _ := first["win"].(bool)
				firstBlood, _ := first["firstBloodKill"].(bool)

				if firstBlood {
					stat, exists := firstBloodStats[champion]
					if !exists {
						stat = &FirstBlood{Champion: champion}
						firstBloodStats[champion] = stat
					}
					stat.MaxGames++
					if win {
						stat.Win++
					}
				}
			}
			exists = append(exists, match)
			totalCalls++
		}
		fmt.Println("Loop Nr:", historyMax)
	}

	if err := saveFirstBloodsToDB(firstBloodStats); err != nil {
		return fmt.Errorf("failed to save first bloods to DB: %v", err)
	}

	return nil
}
